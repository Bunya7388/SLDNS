package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
)

const VERSION = "1.0"

var (
	serverAddr = flag.String("s", "127.0.0.1:53", "DNS server address")
	localAddr  = flag.String("l", "127.0.0.1:8888", "Local listen address")
	verbose    = flag.Bool("v", false, "Verbose logging")
	workers    = flag.Int("w", 4, "Number of worker threads")
)

type Stats struct {
	PacketsSent     uint64
	PacketsReceived uint64
	BytesSent       uint64
	BytesReceived   uint64
}

var stats Stats

type Tunnel struct {
	sessionID  uint32
	serverAddr *net.UDPAddr
	localConn  net.Conn
	udpConn    *net.UDPConn
}

func logf(format string, args ...interface{}) {
	if *verbose {
		log.Printf(format, args...)
	}
}

func genSessionID() uint32 {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return 0
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func buildDNSQuery(data []byte, transactionID uint16) []byte {
	query := make([]byte, 12+len(data))
	query[0] = byte(transactionID >> 8)
	query[1] = byte(transactionID)
	query[2] = 0x01
	query[3] = 0x00
	query[4] = 0x00
	query[5] = 0x01
	query[6] = 0x00
	query[7] = 0x00
	query[8] = 0x00
	query[9] = 0x00
	query[10] = 0x00
	query[11] = 0x00
	copy(query[12:], data)
	return query
}

func parseDNSResponse(response []byte) []byte {
	if len(response) < 12 {
		return nil
	}

	pos := 12
	for pos < len(response) && response[pos] != 0 {
		if (response[pos] & 0xc0) == 0xc0 {
			pos += 2
			break
		}
		pos++
	}
	pos++
	pos += 4

	if pos+10 < len(response) {
		if (response[pos] & 0xc0) == 0xc0 {
			pos += 2
		}
		pos += 8
		rdlen := int(response[pos])<<8 | int(response[pos+1])
		pos += 2
		if pos+rdlen <= len(response) {
			return response[pos : pos+rdlen]
		}
	}

	return nil
}

func (t *Tunnel) sendData(data []byte) error {
	transID := uint16(time.Now().Unix() & 0xFFFF)
	query := buildDNSQuery(data, transID)
	_, err := t.udpConn.WriteToUDP(query, t.serverAddr)
	if err == nil {
		atomic.AddUint64(&stats.PacketsSent, 1)
		atomic.AddUint64(&stats.BytesSent, uint64(len(query)))
	}
	return err
}

func (t *Tunnel) receiveData() ([]byte, error) {
	buffer := make([]byte, 4096)
	t.udpConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := t.udpConn.Read(buffer)
	if err != nil {
		return nil, err
	}

	atomic.AddUint64(&stats.PacketsReceived, 1)
	atomic.AddUint64(&stats.BytesReceived, uint64(n))

	response := parseDNSResponse(buffer[:n])
	if response != nil {
		return response, nil
	}

	return nil, fmt.Errorf("invalid DNS response")
}

func (t *Tunnel) handleConnection() {
	defer t.localConn.Close()

	buffer := make([]byte, 4096)

	for {
		t.localConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		n, err := t.localConn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logf("Read error: %v\n", err)
			}
			return
		}

		if err := t.sendData(buffer[:n]); err != nil {
			logf("Send error: %v\n", err)
			return
		}

		response, err := t.receiveData()
		if err != nil {
			logf("Receive error: %v\n", err)
			continue
		}

		if _, err := t.localConn.Write(response); err != nil {
			logf("Write to client error: %v\n", err)
			return
		}
	}
}

func startServer() {
	listener, err := net.Listen("tcp", *localAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v\n", *localAddr, err)
	}
	defer listener.Close()

	fmt.Printf("DNSTT Client listening on %s\n", *localAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logf("Accept error: %v\n", err)
			continue
		}

		addr, err := net.ResolveUDPAddr("udp", *serverAddr)
		if err != nil {
			logf("Failed to resolve server: %v\n", err)
			conn.Close()
			continue
		}

		udpConn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			logf("Failed to create UDP connection: %v\n", err)
			conn.Close()
			continue
		}

		tunnel := &Tunnel{
			sessionID:  genSessionID(),
			serverAddr: addr,
			localConn:  conn,
			udpConn:    udpConn,
		}

		go tunnel.handleConnection()
	}
}

func statsReporter() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sent := atomic.LoadUint64(&stats.BytesSent)
		received := atomic.LoadUint64(&stats.BytesReceived)
		fmt.Printf("[STATS] Sent: %d bytes | Received: %d bytes | PktSent: %d | PktRecv: %d\n",
			sent, received,
			atomic.LoadUint64(&stats.PacketsSent),
			atomic.LoadUint64(&stats.PacketsReceived))
	}
}

func main() {
	flag.Parse()

	fmt.Printf("=== DNSTT UDP Tunnel Client v%s ===\n", VERSION)
	fmt.Printf("Server: %s\n", *serverAddr)
	fmt.Printf("Local Listen: %s\n", *localAddr)
	fmt.Printf("Workers: %d\n", *workers)

	go statsReporter()
	startServer()
}
