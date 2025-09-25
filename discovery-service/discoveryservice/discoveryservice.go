package discoveryservice

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	multicastPort     = 9999
	heartbeatMessage  = "HEARTBEAT"
	cleanupInterval   = 1 * time.Second
	peerPrintInterval = 3 * time.Second
	peerTimeout       = 10 * time.Second
)

type DiscoveryService struct {
	multicastConn    *net.UDPConn
	multicastAddr    *net.UDPAddr
	peers            map[string]time.Time
	peersMutex       sync.RWMutex
	heartbeatTimeout time.Duration
	cleanupInterval  time.Duration
	stopChan         chan bool
	localIP          string
	nodeID           string
}

func NewDiscoveryService(multicastIP string, heartbeatTimeout time.Duration) (*DiscoveryService, error) {
	networkType := "udp4"
	if isIPv6(multicastIP) {
		networkType = "udp6"
	}
	multicastAddr, err := net.ResolveUDPAddr(networkType, fmt.Sprintf("%s:%d", multicastIP, multicastPort))
	if err != nil {
		return nil, fmt.Errorf("error resolving multicast address: %s", err)
	}

	conn, err := net.ListenMulticastUDP(networkType, nil, multicastAddr)
	if err != nil {
		return nil, fmt.Errorf("error creating multicast connection: %s", err)
	}

	localIP := os.Getenv("DOCKER_IP")

	nodeID := os.Getenv("NODE_ID")

	fmt.Printf("[Node %s] Multicast listening on %s:%d (%s)\n", nodeID, multicastIP, multicastPort, networkType)
	fmt.Printf("[Node %s] Local IP: %s\n", nodeID, localIP)

	return &DiscoveryService{
		multicastConn:    conn,
		multicastAddr:    multicastAddr,
		peers:            make(map[string]time.Time),
		heartbeatTimeout: heartbeatTimeout,
		cleanupInterval:  cleanupInterval,
		stopChan:         make(chan bool),
		localIP:          localIP,
		nodeID:           nodeID,
	}, nil
}

func (ds *DiscoveryService) Start() {
	fmt.Printf("[Node %s] Starting discovery service\n", ds.nodeID)
	go ds.Listen()
	go ds.SendHeartbeats()
	go ds.PrintPeers()
	go ds.CleanUpPeers()
}

func (ds *DiscoveryService) Stop() {
	fmt.Printf("[Node %s] Stopping discovery service\n", ds.nodeID)
	close(ds.stopChan)
	ds.multicastConn.Close()
}

func (ds *DiscoveryService) Listen() {
	buf := make([]byte, 1024)
	fmt.Printf("[Node %s] Listening for multicast packets on %s\n", ds.nodeID, ds.multicastAddr.String())

	for {
		select {
		case <-ds.stopChan:
			fmt.Printf("[Node %s] Stopping listener\n", ds.nodeID)
			return
		default:
			n, src, err := ds.multicastConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Printf("[Node %s] Error reading from UDP: %s\n", ds.nodeID, err)
				continue
			}

			fmt.Printf("[Node %s] Received %d bytes from %s\n", ds.nodeID, n, src.IP.String())

			if string(buf[:n]) == heartbeatMessage {
				fmt.Printf("[Node %s] Heartbeat from %s\n", ds.nodeID, src.IP.String())
				ds.updatePeer(src.IP.String())
			} else {
				fmt.Printf("[Node %s] Unknown message from %s: %s\n", ds.nodeID, src.IP.String(), string(buf[:n]))
			}
		}
	}
}

func (ds *DiscoveryService) SendHeartbeats() {
	ticker := time.NewTicker(ds.heartbeatTimeout)
	defer ticker.Stop()

	fmt.Printf("[Node %s] Sending heartbeats every %v to %s\n", ds.nodeID, ds.heartbeatTimeout, ds.multicastAddr.String())

	for {
		select {
		case <-ds.stopChan:
			fmt.Printf("[Node %s] Stopping heartbeat sender\n", ds.nodeID)
			return
		case <-ticker.C:
			n, err := ds.multicastConn.WriteToUDP([]byte(heartbeatMessage), ds.multicastAddr)
			if err != nil {
				fmt.Printf("[Node %s] Error sending heartbeat: %s\n", ds.nodeID, err)
			} else {
				fmt.Printf("[Node %s] Sent heartbeat (%d bytes)\n", ds.nodeID, n)
			}
		}
	}
}

func (ds *DiscoveryService) CleanUpPeers() {
	ticker := time.NewTicker(ds.cleanupInterval)
	defer ticker.Stop()

	fmt.Printf("[Node %s] Starting peer cleanup every %v\n", ds.nodeID, ds.cleanupInterval)

	for {
		select {
		case <-ds.stopChan:
			fmt.Printf("[Node %s] Stopping peer cleanup\n", ds.nodeID)
			return
		case <-ticker.C:
			ds.peersMutex.Lock()
			now := time.Now()
			for ip, lastSeen := range ds.peers {
				if now.Sub(lastSeen) > peerTimeout {
					delete(ds.peers, ip)
					fmt.Printf("[Node %s] Removed inactive peer: %s\n", ds.nodeID, ip)
				}
			}
			ds.peersMutex.Unlock()
		}
	}
}

func (ds *DiscoveryService) updatePeer(ip string) {
	ds.peersMutex.Lock()
	defer ds.peersMutex.Unlock()

	if ip == ds.localIP {
		return
	}

	ds.peers[ip] = time.Now()
	fmt.Printf("[Node %s] Updated peer: %s (total peers: %d)\n", ds.nodeID, ip, len(ds.peers))
}

func (ds *DiscoveryService) PrintPeers() {
	ticker := time.NewTicker(peerPrintInterval)
	defer ticker.Stop()

	fmt.Printf("[Node %s] Starting peer printer every %v\n", ds.nodeID, peerPrintInterval)

	for {
		select {
		case <-ds.stopChan:
			fmt.Printf("[Node %s] Stopping peer printer\n", ds.nodeID)
			return
		case <-ticker.C:
			ds.peersMutex.RLock()
			now := time.Now()

			if len(ds.peers) == 0 {
				fmt.Printf("[Node %s] No active peers found\n", ds.nodeID)
			} else {
				fmt.Printf("[Node %s] Active peers (%d):\n", ds.nodeID, len(ds.peers))
				for ip, lastSeen := range ds.peers {
					age := now.Sub(lastSeen).Round(time.Millisecond)
					fmt.Printf("[Node %s]  %s (last seen: %v ago)\n", ds.nodeID, ip, age)
				}
			}
			ds.peersMutex.RUnlock()
		}
	}
}

func isIPv6(addr string) bool {
	ip := net.ParseIP(addr)
	return ip != nil && ip.To16() != nil && ip.To4() == nil
}
