package network

import (
	"log"
	"net"
	proto "snake-game/internal/proto/gen"
)

type Manager struct {
	unicastConn   *net.UDPConn
	multicastConn *net.UDPConn
	role          proto.NodeRole
	msgSeq        int64
}

func NewNetworkManager() *Manager {
	return &Manager{
		role:   proto.NodeRole_NORMAL,
		msgSeq: 1,
	}
}

func (m *Manager) Start() error {
	if err := m.setupUnicastSocket(); err != nil {
		return err
	}
	if err := m.setupMulticastSocket(); err != nil {
		return err
	}
	go m.listenForMessages()
	return nil
}

func (m *Manager) setupUnicastSocket() error {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	m.unicastConn = conn
	log.Printf("Unicast socket bound to %s", conn.LocalAddr())
	return nil
}

func (m *Manager) setupMulticastSocket() error {
	groupAddr, err := net.ResolveUDPAddr("udp", "239.192.0.4:9192")
	if err != nil {
		return err
	}
	conn, err := net.ListenMulticastUDP("udp", nil, groupAddr)
	if err != nil {
		return err
	}
	m.multicastConn = conn
	log.Printf("Multicast socket joined group %s", groupAddr)
	return nil
}

func (m *Manager) listenForMessages() {
	buf := make([]byte, 4096)
	for {
		n, addr, err := m.unicastConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}
		go m.handleMessage(buf[:n], addr)
	}
}
