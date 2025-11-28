package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-serial-tty-client/client"
	"go-serial-tty-client/serial"
)

func main() {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-device"
	}

	id := flag.String("id", hostname, "Device ID")
	addr := flag.String("addr", "localhost:80", "Server address (host:port)")
	portPath := flag.String("port", "/dev/ttyUSB0", "Serial port path")
	baud := flag.Int("baud", 115200, "Baud rate")
	sim := flag.Bool("sim", false, "Enable simulation mode")
	debugWS := flag.Bool("debug-ws", false, "Enable WebSocket debug logging")
	debugSerial := flag.Bool("debug-serial", false, "Enable Serial debug logging")
	appendNewline := flag.Bool("append-newline", false, "Append \\r\\n to incoming WebSocket messages")
	flag.Parse()

	log.Printf("Starting Serial TTY Client (ID: %s, Server: %s)", *id, *addr)

	var p serial.Port
	var err error

	if *sim {
		log.Println("Mode: Simulation")
		p = serial.NewSimulatedPort()
	} else {
		log.Printf("Mode: Real Serial (%s @ %d)", *portPath, *baud)
		p, err = serial.NewRealPort(*portPath, *baud)
		if err != nil {
			log.Fatalf("Failed to open serial port: %v", err)
		}
	}

	if *debugSerial {
		p = serial.NewDebugPort(p)
	}

	wsClient := client.NewWSClient(*addr, *id, p, *debugWS, *appendNewline)
	if err := wsClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to WebSocket server: %v", err)
	}

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down...")
		// In a real app we might want to close connections gracefully here
		// But for now we rely on the main function exiting or the client logic handling it
		os.Exit(0)
	}()

	log.Println("Connected. Forwarding data...")
	wsClient.Start()
}
