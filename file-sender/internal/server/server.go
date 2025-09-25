package server

import (
	"context"
	"file-sender/internal/protocol"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type Server struct {
	port     string
	stopChan chan struct{}
	wg       sync.WaitGroup
}

type clientStats struct {
	startTime      time.Time
	lastUpdateTime time.Time
	totalBytes     int64
	lastBytes      int64
}

const (
	filesDir = "uploads"
)

func NewServer(port string) *Server {
	return &Server{
		port:     port,
		stopChan: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	os.Mkdir(filesDir, 0777)
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("error starting s: %s", err)
	}
	defer listener.Close()
	fmt.Println("Listening on " + ":" + s.port)

	go func() {
		<-s.stopChan
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return nil
			default:
				return fmt.Errorf("error accepting connection: %s", err)
			}
		}
		go func() {
			err := s.handleConnection(conn)
			if err != nil {
				fmt.Printf("error handling connection: %s", err)
			}
		}()
	}
}

func (s *Server) Stop(ctx context.Context) error {
	close(s.stopChan)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) handleConnection(conn net.Conn) error {
	s.wg.Add(1)
	defer s.wg.Done()
	defer conn.Close()
	stats := &clientStats{
		startTime:      time.Now(),
		lastUpdateTime: time.Now(),
	}

	filename, fileSize, err := protocol.ReadFileMetadata(conn)
	if err != nil {
		return fmt.Errorf("error reading file metadata: %s", err)
	}

	file, err := os.Create(filesDir + "/" + filename)
	if err != nil {
		return fmt.Errorf("error creating file: %s", err)
	}
	defer file.Close()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.printSpeedStats(conn.RemoteAddr().String(), stats)
			case <-done:
				return
			}
		}
	}()

	reader := &progressReader{conn: conn, stats: stats}
	n, err := io.CopyN(file, reader, fileSize)
	if err != nil {
		close(done)
		file.Close()
		os.Remove(filesDir + "/" + filename)
		return fmt.Errorf("error writing to file: %s", err)
	}
	close(done)

	s.printSpeedStats(conn.RemoteAddr().String(), stats)

	success := n == fileSize
	if success {
		fmt.Printf("Successfully received %d bytes and saved to %s\n", n, filesDir+"/"+filename)
		conn.Write([]byte{1})
	} else {
		fmt.Printf("Error: received %d bytes but expected %d\n", n, fileSize)
		conn.Write([]byte{0})
		os.Remove(filesDir + "/" + filename)
	}

	return nil
}

func (s *Server) printSpeedStats(clientAddr string, stats *clientStats) {
	now := time.Now()
	elapsed := now.Sub(stats.lastUpdateTime).Seconds()
	totalElapsed := now.Sub(stats.startTime).Seconds()

	if elapsed > 0 {
		instantSpeed := float64(stats.totalBytes-stats.lastBytes) / elapsed
		averageSpeed := float64(stats.totalBytes) / totalElapsed

		fmt.Printf("Client %s: Instant speed: %.2f B/s, Average speed: %.2f B/s, Total: %d bytes\n",
			clientAddr, instantSpeed, averageSpeed, stats.totalBytes)
	}

	stats.lastUpdateTime = now
	stats.lastBytes = stats.totalBytes
}

type progressReader struct {
	conn  net.Conn
	stats *clientStats
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.conn.Read(p)
	r.stats.totalBytes += int64(n)
	return n, err
}
