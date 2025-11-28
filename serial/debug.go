package serial

import (
	"log"
)

// DebugPort wraps a Port and logs all Read/Write operations
type DebugPort struct {
	Port
}

func NewDebugPort(p Port) *DebugPort {
	return &DebugPort{Port: p}
}

func (d *DebugPort) Read(b []byte) (int, error) {
	n, err := d.Port.Read(b)
	if n > 0 {
		log.Printf("[SERIAL DEBUG] Read %d bytes: %q", n, b[:n])
	}
	if err != nil {
		log.Printf("[SERIAL DEBUG] Read error: %v", err)
	}
	return n, err
}

func (d *DebugPort) Write(b []byte) (int, error) {
	log.Printf("[SERIAL DEBUG] Write %d bytes: %q", len(b), b)
	return d.Port.Write(b)
}
