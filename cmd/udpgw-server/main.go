package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const VERSION = "1.0"

var (
	listenAddr = flag.String("l", "0.0.0.0:5555", "Listen address:port")
	remoteAddr = flag.String("r", "127.0.0.1:53", "Remote address")
	workers    = flag.Int("w", 8, "Worker threads")
	verbose    = flag.Bool("v", false, "Verbose")
	timeout    = flag.Int("t", 300, "Timeout seconds")
)

type Stats struct {
	PacketsRecv uint64
	PacketsSent uint64
	BytesRecv   uint64
	BytesSent   uint64
	Sessions    uint64
}

type Session struct {
	ID   string
	Conn *net.UDPConn
	Addr *net.UDPAddr
	Seen time.Time
	Lock sync.Mutex
}

type Gateway struct {
	listener *net.UDPConn
	remote   *net.UDPAddr
	sessions map[string]*Session
	lock     sync.RWMutex
	queue    chan *UDPPacket
	done     chan struct{}
	wg       sync.WaitGroup
}

type UDPPacket struct {
	Data []byte
	Addr *net.UDPAddr
}

var stats Stats

func logf(msg string, args ...interface{}) {
	if *verbose {
		log.Printf(msg, args...)
	}
}

func newGateway(listen, remote string) (*Gateway, error) {
	l, err := net.ResolveUDPAddr("udp", listen)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", l)
	if err != nil {
		return nil, err
	}
	r, err := net.ResolveUDPAddr("udp", remote)
	if err != nil {
		return nil, err
	}
	conn.SetReadBuffer(262144)
	conn.SetWriteBuffer(262144)
	return &Gateway{
		listener: conn,
		remote:   r,
		sessions: make(map[string]*Session),
		queue:    make(chan *UDPPacket, *workers*5),
		done:     make(chan struct{}),
	}, nil
}

func (g *Gateway) getID(addr *net.UDPAddr) string {
	return addr.String()
}

func (g *Gateway) getSession(id string, addr *net.UDPAddr) (*Session, error) {
	g.lock.RLock()
	if s, ok := g.sessions[id]; ok {
		s.Lock.Lock()
		s.Seen = time.Now()
		s.Lock.Unlock()
		g.lock.RUnlock()
		return s, nil
	}
	g.lock.RUnlock()

	c, err := net.DialUDP("udp", nil, g.remote)
	if err != nil {
		return nil, err
	}
	s := &Session{ID: id, Conn: c, Addr: addr, Seen: time.Now()}
	g.lock.Lock()
	g.sessions[id] = s
	g.lock.Unlock()
	atomic.AddUint64(&stats.Sessions, 1)
	go g.readRemote(s)
	return s, nil
}

func (g *Gateway) readRemote(s *Session) {
	defer func() {
		s.Conn.Close()
		g.lock.Lock()
		delete(g.sessions, s.ID)
		g.lock.Unlock()
		atomic.AddUint64(&stats.Sessions, ^uint64(0))
	}()
	buf := make([]byte, 65535)
	for {
		select {
		case <-g.done:
			return
		default:
		}
		s.Conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		n, err := s.Conn.Read(buf)
		if err != nil {
			return
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		g.listener.WriteToUDP(data, s.Addr)
		atomic.AddUint64(&stats.PacketsSent, 1)
		atomic.AddUint64(&stats.BytesSent, uint64(n))
	}
}

func (g *Gateway) process(pkt *UDPPacket) {
	id := g.getID(pkt.Addr)
	s, err := g.getSession(id, pkt.Addr)
	if err != nil {
		logf("Error: %v\n", err)
		return
	}
	s.Conn.Write(pkt.Data)
	atomic.AddUint64(&stats.PacketsRecv, 1)
	atomic.AddUint64(&stats.BytesRecv, uint64(len(pkt.Data)))
}

func (g *Gateway) worker() {
	defer g.wg.Done()
	for {
		select {
		case pkt := <-g.queue:
			if pkt == nil {
				return
			}
			g.process(pkt)
		case <-g.done:
			return
		}
	}
}

func (g *Gateway) start() {
	for i := 0; i < *workers; i++ {
		g.wg.Add(1)
		go g.worker()
	}
	g.wg.Add(1)
	go g.read()
	g.wg.Add(1)
	go g.cleanup()
}

func (g *Gateway) read() {
	defer g.wg.Done()
	buf := make([]byte, 65535)
	for {
		g.listener.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, addr, err := g.listener.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		select {
		case g.queue <- &UDPPacket{data, addr}:
		case <-g.done:
			return
		}
	}
}

func (g *Gateway) cleanup() {
	defer g.wg.Done()
	tick := time.NewTicker(1 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			g.lock.Lock()
			now := time.Now()
			td := time.Duration(*timeout) * time.Second
			for id, s := range g.sessions {
				s.Lock.Lock()
				if now.Sub(s.Seen) > td {
					s.Conn.Close()
					delete(g.sessions, id)
					atomic.AddUint64(&stats.Sessions, ^uint64(0))
				}
				s.Lock.Unlock()
			}
			g.lock.Unlock()
		case <-g.done:
			return
		}
	}
}

func (g *Gateway) stop() {
	close(g.done)
	g.listener.Close()
	g.wg.Wait()
	g.lock.Lock()
	for _, s := range g.sessions {
		s.Conn.Close()
	}
	g.lock.Unlock()
}

func main() {
	flag.Parse()
	if *workers < 1 {
		*workers = 1
	}
	if *workers > 256 {
		*workers = 256
	}

	fmt.Printf("=== UDPGW UDP Gateway v%s ===\n", VERSION)
	fmt.Printf("Listen: %s\n", *listenAddr)
	fmt.Printf("Remote: %s\n", *remoteAddr)
	fmt.Printf("Workers: %d\n", *workers)

	gw, err := newGateway(*listenAddr, *remoteAddr)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	defer gw.stop()

	gw.start()

	go func() {
		tick := time.NewTicker(30 * time.Second)
		defer tick.Stop()
		for range tick.C {
			fmt.Printf("[STATS] Recv: %d | Sent: %d | Bytes In: %.1f MB | Out: %.1f MB | Sessions: %d\n",
				atomic.LoadUint64(&stats.PacketsRecv),
				atomic.LoadUint64(&stats.PacketsSent),
				float64(atomic.LoadUint64(&stats.BytesRecv))/1e6,
				float64(atomic.LoadUint64(&stats.BytesSent))/1e6,
				atomic.LoadUint64(&stats.Sessions))
		}
	}()

	fmt.Println("[*] Running...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\n[*] Shutting down...")
}
