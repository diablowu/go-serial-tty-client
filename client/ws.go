package client

import (
	"log"
	"net/url"

	"go-serial-tty-client/serial"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	serverAddr    string
	deviceID      string
	serialPort    serial.Port
	conn          *websocket.Conn
	done          chan struct{}
	debug         bool
	appendNewline bool
}

func NewWSClient(serverAddr, deviceID string, port serial.Port, debug, appendNewline bool) *WSClient {
	return &WSClient{
		serverAddr:    serverAddr,
		deviceID:      deviceID,
		serialPort:    port,
		done:          make(chan struct{}),
		debug:         debug,
		appendNewline: appendNewline,
	}
}

func (c *WSClient) Connect() error {
	u := url.URL{Scheme: "ws", Host: c.serverAddr, Path: "/ws/device", RawQuery: "id=" + c.deviceID}
	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *WSClient) Start() {
	defer c.conn.Close()
	defer c.serialPort.Close()

	// Read from WS and write to Serial
	go func() {
		defer close(c.done)
		for {
			mt, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Println("WS read error:", err)
				return
			}
			if c.debug {
				log.Printf("[WS DEBUG] Read message (type %d): %q", mt, message)
			}

			// If binary, write directly to serial
			// If text, we might want to log it or write it too depending on protocol
			// The reference implementation echoes binary back and prefixes text.
			// Here we want to forward to serial.

			if mt == websocket.BinaryMessage || mt == websocket.TextMessage {
				payload := message
				if c.appendNewline {
					payload = append(payload, []byte("\r\n")...)
				}

				if c.debug {
					log.Printf("[WS DEBUG] Forwarding %d bytes to serial", len(payload))
				}
				_, err := c.serialPort.Write(payload)
				if err != nil {
					log.Println("Serial write error:", err)
					return
				}
			} else {
				if c.debug {
					log.Printf("[WS DEBUG] Ignoring message type %d", mt)
				}
			}
		}
	}()

	// Read from Serial and write to WS
	buffer := make([]byte, 1024)
	for {
		select {
		case <-c.done:
			return
		default:
			n, err := c.serialPort.Read(buffer)
			if err != nil {
				log.Println("Serial read error:", err)
				return
			}
			if n > 0 {
				// Send as binary or text? Reference sends TextMessage for serial output simulation
				// But usually serial data is raw bytes. Let's send as Binary for correctness,
				// or Text if it's expected to be human readable.
				// The reference simulator sends `websocket.TextMessage` for simulated input.
				// Let's stick to TextMessage for now to match the simulator's behavior for "console" like output,
				// but strictly speaking serial is binary.
				// However, the user said "serial reading should be carefully optimized".
				// Let's send as BinaryMessage to be safe for all data types, unless the server expects Text.
				// Looking at reference:
				// Simulator sends: `c.WriteMessage(websocket.TextMessage, []byte(text+"\r\n"))`
				// Simulator echoes binary as binary.
				// Let's assume the server handles both. I'll send as Binary for raw serial data.
				// Wait, if I want to see it in a web terminal, it might expect text.
				// Let's try to detect or just send Binary.
				// Actually, for a TTY, it's usually text.
				// Let's send as BinaryMessage, it's more generic.

				if c.debug {
					log.Printf("[WS DEBUG] Writing message: %q", buffer[:n])
				}
				err = c.conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
				if err != nil {
					log.Println("WS write error:", err)
					return
				}
			}
		}
	}
}
