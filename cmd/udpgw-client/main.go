package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const VERSION = "1.0"

var (
	listenAddr = flag.String("l", "127.0.0.1:5555", "Local listen")
	remoteAddr = flag.String("r", "127.0.0.1:5555", "Remote gateway")
	verbose    = flag.Bool("v", false, "Verbose")
)

type Stats struct {
	RecvPkts uint64
	SentPkts uint64
	RecvByte uint64
	SentByte uint64
}

var stats Stats

func logf(msg string, args ...interface{}) {
	if *verbose {
		log.Printf(msg, args...)
	}
}

func main() {
	flag.Parse()

	fmt.Printf("=== UDPGW Client v%s ===\n", VERSION)
	fmt.Printf("Listen: %s\n", *listenAddr)
	fmt.Printf("Remote: %s\n", *remoteAddr)

	listenUDP, err := net.ResolveUDPAddr("udp", *listenAddr)
	if err != nil {
		log.Fatalf("Invalid listen: %v\n", err)
	}
	listener, err := net.ListenUDP("udp", listenUDP)
	if err != nil {
		log.Fatalf("Listen error: %v\n", err)
	}
	defer listener.Close()

	listener.SetReadBuffer(262144)
	listener.SetWriteBuffer(262144)

	remoteUDP, err := net.ResolveUDPAddr("udp", *remoteAddr)
	if err != nil {
		log.Fatalf("Invalid remote: %v\n", err)
	}
	remote, err := net.DialUDP("udp", nil, remoteUDP)
	if err != nil {
		log.Fatalf("Dial error: %v\n", err)
	}
	defer remote.Close()

	remote.SetReadBuffer(262144)
	remote.SetWriteBuffer(262144)

	fmt.Println("[*] Running...")

	// Read from remote and relay
	go func() {
		buf := make([]byte, 65535)
		for {
			n, err := remote.Read(buf)
			if err != nil {
				logf("Remote read: %v\n", err)
				return
			}
			atomic.AddUint64(&stats.SentPkts, 1)
			atomic.AddUint64(&stats.SentByte, uint64(n))
		}
	}()

	// Stats
	go func() {
		tick := time.NewTicker(30 * time.Second)
		defer tick.Stop()
		for range tick.C {
			fmt.Printf("[STATS] In: %d pkts | Out: %d pkts | Recv: %.1f MB | Sent: %.1f MB\n",
				atomic.LoadUint64(&stats.RecvPkts),
				atomic.LoadUint64(&stats.SentPkts),
				float64(atomic.LoadUint64(&stats.RecvByte))/1e6,
				float64(atomic.LoadUint64(&stats.SentByte))/1e6)
		}
	}()

	// Main loop
	buf := make([]byte, 65535)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool)
	go func() {
		<-sig
		fmt.Println("\n[*] Shutdown...")
		done <- true
	}()

	for {
		select {
		case <-done:
			return
		default:
		}
		listener.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := listener.ReadFromUDP(buf)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				continue
			}
			logf("Read: %v\n", err)
			continue
		}
		_, err = remote.Write(buf[:n])
		if err != nil {
			logf("Write: %v\n", err)
			continue
		}
		atomic.AddUint64(&stats.RecvPkts, 1)
		atomic.AddUint64(&stats.RecvByte, uint64(n))
		logf("Fwd %d bytes from %s\n", n, addr)
	}
}
