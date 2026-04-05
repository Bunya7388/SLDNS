package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	MAX_DNS_PAYLOAD = 512
	CHUNK_SIZE      = 240
	VERSION         = "1.0"
)

var (
	listenAddr    = flag.String("l", "0.0.0.0:53", "Listen address:port")
	workers       = flag.Int("h", 4, "Number of worker threads")
	verbose       = flag.Bool("v", false, "Verbose logging")
	logFile       = flag.String("log", "", "Log file path")
	statsInterval = flag.Int("stats", 30, "Stats reporting interval (seconds)")
)

// SessionManager maintains active tunnel sessions
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uint32]*Session
}

// Session represents a tunnel session
type Session struct {
	ID         uint32
	Created    time.Time
	LastActive time.Time
	Buffer     *bytes.Buffer
	Lock       sync.Mutex
	ChunkCount uint64
}

// Stats tracks performance metrics
type Stats struct {
	TotalPackets   uint64
	TotalBytes     uint64
	ActiveSessions uint64
	PacketsPerSec  float64
	BytesPerSec    float64
}

var (
	sessionMgr *SessionManager
	stats      Stats
	statsLock  sync.RWMutex
)

// DNSRequest wraps a DNS query with client info
type DNSRequest struct {
	Query  []byte
	Addr   *net.UDPAddr
	Server *net.UDPConn
}

// WorkerPool processes DNS requests concurrently
type WorkerPool struct {
	workers int
	queue   chan *DNSRequest
	done    chan struct{}
	wg      sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	wp := &WorkerPool{
		workers: numWorkers,
		queue:   make(chan *DNSRequest, numWorkers*2),
		done:    make(chan struct{}),
	}
	return wp
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
	logf("Started %d workers\n", wp.workers)
}

func (wp *WorkerPool) Stop() {
	close(wp.done)
	wp.wg.Wait()
	logf("All workers stopped\n")
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case req := <-wp.queue:
			if req == nil {
				return
			}
			processDNSQuery(req)
		case <-wp.done:
			return
		}
	}
}

func (wp *WorkerPool) Submit(req *DNSRequest) {
	select {
	case wp.queue <- req:
	case <-wp.done:
		logf("Worker pool stopped, dropping request\n")
	}
}

func logf(format string, args ...interface{}) {
	if *verbose {
		log.Printf(format, args...)
	}
}

func (sm *SessionManager) GetOrCreate(sessionID uint32) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActive = time.Now()
		return session
	}

	session := &Session{
		ID:         sessionID,
		Created:    time.Now(),
		LastActive: time.Now(),
		Buffer:     bytes.NewBuffer(make([]byte, 0, MAX_DNS_PAYLOAD*10)),
	}
	sm.sessions[sessionID] = session
	atomic.AddUint64(&stats.ActiveSessions, 1)
	return session
}

func (sm *SessionManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	toDelete := []uint32{}
	for id, session := range sm.sessions {
		if now.Sub(session.LastActive) > 5*time.Minute {
			toDelete = append(toDelete, id)
		}
	}
	for _, id := range toDelete {
		delete(sm.sessions, id)
		atomic.AddUint64(&stats.ActiveSessions, ^uint64(0))
	}
}

func encodeData(data []byte, sessionID uint32) []byte {
	header := make([]byte, 8)
	binary.BigEndian.PutUint32(header[0:4], sessionID)
	binary.BigEndian.PutUint32(header[4:8], uint32(len(data)))
	return append(header, data...)
}

func decodeData(packet []byte) (uint32, []byte, error) {
	if len(packet) < 8 {
		return 0, nil, fmt.Errorf("packet too small")
	}
	sessionID := binary.BigEndian.Uint32(packet[0:4])
	length := binary.BigEndian.Uint32(packet[4:8])
	if len(packet) < 8+int(length) {
		return 0, nil, fmt.Errorf("incomplete data")
	}
	return sessionID, packet[8 : 8+length], nil
}

func buildDNSResponse(query []byte, sessionID uint32) []byte {
	if len(query) < 12 {
		return query
	}

	response := make([]byte, len(query))
	copy(response, query)

	// Set response flags: QR=1, AA=1, TC=0, RD=1, RA=1, RCODE=0
	response[2] = 0x84
	response[3] = 0x00
	// QDCOUNT and ANCOUNT
	binary.BigEndian.PutUint16(response[4:6], 1)
	binary.BigEndian.PutUint16(response[6:8], 1)

	// Extract question
	qPos := 12
	for qPos < len(response)-4 {
		if response[qPos] == 0 {
			qPos++
			break
		}
		if (response[qPos] & 0xc0) == 0xc0 {
			qPos += 2
			break
		}
		qPos++
	}
	qPos += 4 // Skip QTYPE and QCLASS

	// Add answer record
	if qPos+14 < len(response) {
		copy(response[qPos:], query[12:])
		copy(response[qPos:], []byte{0xc0, 0x0c})
		binary.BigEndian.PutUint16(response[qPos+2:], 1)  // TYPE A
		binary.BigEndian.PutUint16(response[qPos+4:], 1)  // CLASS IN
		binary.BigEndian.PutUint32(response[qPos+6:], 60) // TTL
		binary.BigEndian.PutUint16(response[qPos+10:], 4) // RDLENGTH
		response[qPos+12] = byte(sessionID >> 24)
		response[qPos+13] = byte(sessionID >> 16)
		response[qPos+14] = byte(sessionID >> 8)
		response[qPos+15] = byte(sessionID)
	}

	return response
}

func processDNSQuery(req *DNSRequest) {
	defer func() {
		if r := recover(); r != nil {
			logf("Panic in processDNSQuery: %v\n", r)
		}
	}()

	query := req.Query
	if len(query) < 12 {
		return
	}

	// Generate session ID from query
	h := fnv.New32a()
	h.Write(query[12:])
	sessionID := h.Sum32()

	session := sessionMgr.GetOrCreate(sessionID)

	// Build response
	response := buildDNSResponse(query, sessionID)

	// Send response
	_, err := req.Server.WriteToUDP(response, req.Addr)
	if err != nil {
		logf("Error writing DNS response: %v\n", err)
		return
	}

	// Update stats
	atomic.AddUint64(&stats.TotalPackets, 1)
	atomic.AddUint64(&stats.TotalBytes, uint64(len(query)+len(response)))
	session.ChunkCount++
}

func startStatsReporter() {
	ticker := time.NewTicker(time.Duration(*statsInterval) * time.Second)
	lastPackets := uint64(0)
	lastBytes := uint64(0)
	lastTime := time.Now()

	go func() {
		for range ticker.C {
			now := time.Now()
			currentPackets := atomic.LoadUint64(&stats.TotalPackets)
			currentBytes := atomic.LoadUint64(&stats.TotalBytes)
			activeSessions := atomic.LoadUint64(&stats.ActiveSessions)

			elapsed := now.Sub(lastTime).Seconds()
			pps := float64(currentPackets-lastPackets) / elapsed
			bps := float64(currentBytes-lastBytes) / elapsed

			fmt.Printf("[STATS] Packets: %d | PPS: %.0f | Bytes: %d | BPS: %.0f Mbps | Sessions: %d\n",
				currentPackets, pps, currentBytes, bps/1e6, activeSessions)

			lastPackets = currentPackets
			lastBytes = currentBytes
			lastTime = now
		}
	}()
}

func handleSignals(wp *WorkerPool) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[INFO] Shutting down...")
		wp.Stop()
		sessionMgr.Cleanup()
		os.Exit(0)
	}()
}

func init() {
	sessionMgr = &SessionManager{
		sessions: make(map[uint32]*Session),
	}
}

func main() {
	flag.Parse()

	if *workers < 1 {
		*workers = 1
	}
	if *workers > 256 {
		*workers = 256
	}

	// Setup logging
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Cannot open log file: %v\n", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	fmt.Printf("=== DNSTT UDP Tunnel Server v%s ===\n", VERSION)
	fmt.Printf("Workers: %d\n", *workers)
	fmt.Printf("Listen: %s\n", *listenAddr)
	fmt.Printf("Verbose: %v\n", *verbose)

	wp := NewWorkerPool(*workers)
	wp.Start()

	startStatsReporter()
	handleSignals(wp)

	// Listen for UDP packets
	udpAddr, err := net.ResolveUDPAddr("udp", *listenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve address: %v\n", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}
	defer conn.Close()

	if err := conn.SetReadBuffer(4 * 1024 * 1024); err != nil {
		logf("Failed to set read buffer: %v\n", err)
	}
	if err := conn.SetWriteBuffer(4 * 1024 * 1024); err != nil {
		logf("Failed to set write buffer: %v\n", err)
	}

	fmt.Printf("DNS Server listening on %s with %d workers\n", *listenAddr, *workers)

	buffer := make([]byte, 4096)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logf("Error reading from UDP: %v\n", err)
			continue
		}

		query := make([]byte, n)
		copy(query, buffer[:n])

		req := &DNSRequest{
			Query:  query,
			Addr:   remoteAddr,
			Server: conn,
		}

		wp.Submit(req)
	}
}
