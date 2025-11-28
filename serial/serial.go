package serial

import (
	"errors"
	"log"
	"time"

	"go.bug.st/serial"
)

// Port interface defines the methods we need for serial communication
type Port interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

// RealPort implements Port using a real serial device
type RealPort struct {
	port serial.Port
}

func NewRealPort(portName string, baudRate int) (*RealPort, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	return &RealPort{port: port}, nil
}

func (p *RealPort) Read(b []byte) (int, error) {
	return p.port.Read(b)
}

func (p *RealPort) Write(b []byte) (int, error) {
	return p.port.Write(b)
}

func (p *RealPort) Close() error {
	return p.port.Close()
}

// SimulatedPort implements Port for testing without hardware
type SimulatedPort struct {
	readChan chan []byte
	close    chan struct{}
}

func NewSimulatedPort() *SimulatedPort {
	s := &SimulatedPort{
		readChan: make(chan []byte, 10),
		close:    make(chan struct{}),
	}
	go s.simulateInput()
	return s
}

func (s *SimulatedPort) simulateInput() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-s.close:
			return
		case <-ticker.C:
			counter++
			msg := []byte("Simulated Serial Output " + time.Now().Format(time.RFC3339) + "\r\n")
			select {
			case s.readChan <- msg:
			default:
				// Drop if buffer full
			}
		}
	}
}

func (s *SimulatedPort) Read(b []byte) (int, error) {
	select {
	case <-s.close:
		return 0, errors.New("port closed")
	case data := <-s.readChan:
		copy(b, data)
		return len(data), nil
	}
}

func (s *SimulatedPort) Write(b []byte) (int, error) {
	log.Printf("[SIM] Serial Write: %s", string(b))
	return len(b), nil
}

func (s *SimulatedPort) Close() error {
	close(s.close)
	return nil
}
