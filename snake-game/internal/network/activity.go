package network

import (
	"net"
	"sync"
	"time"
)

type ActivityManager struct {
	mu            sync.RWMutex
	lastSent      map[string]time.Time
	lastRecv      map[string]time.Time
	stateDelayMs  int32
	pingTicker    *time.Ticker
	timeoutTicker *time.Ticker
	manager       *Manager
}

func NewActivityManager(stateDelayMs int32, manager *Manager) *ActivityManager {
	am := &ActivityManager{
		lastSent:     make(map[string]time.Time),
		lastRecv:     make(map[string]time.Time),
		stateDelayMs: stateDelayMs,
		manager:      manager,
	}
	am.startMonitoring()
	return am
}

func (am *ActivityManager) startMonitoring() {
	pingInterval := time.Duration(am.stateDelayMs) * time.Millisecond / 10
	am.pingTicker = time.NewTicker(pingInterval)

	timeoutInterval := time.Duration(am.stateDelayMs) * time.Millisecond / 5
	am.timeoutTicker = time.NewTicker(timeoutInterval)

	go am.monitorActivity()
}

func (am *ActivityManager) monitorActivity() {
	for {
		select {
		case <-am.pingTicker.C:
			am.checkAndSendPings()
		case <-am.timeoutTicker.C:
			am.checkTimeouts()
		case <-am.manager.closeChan:
			return
		}
	}
}

func (am *ActivityManager) checkAndSendPings() {
	am.mu.RLock()
	defer am.mu.RUnlock()

	now := time.Now()
	pingThreshold := time.Duration(am.stateDelayMs) * time.Millisecond / 10

	for addrStr, lastSent := range am.lastSent {
		if now.Sub(lastSent) > pingThreshold {
			addr, err := net.ResolveUDPAddr("udp", addrStr)
			if err == nil {
				go am.manager.sendPing(addr)
				go am.RecordMessageSent(addr)
			}
		}
	}
}

func (am *ActivityManager) checkTimeouts() {
	am.mu.RLock()
	defer am.mu.RUnlock()

	now := time.Now()
	timeoutThreshold := time.Duration(float64(am.stateDelayMs)*0.8) * time.Millisecond

	for addrStr, lastRecv := range am.lastRecv {
		if now.Sub(lastRecv) > timeoutThreshold {
			addr, err := net.ResolveUDPAddr("udp", addrStr)
			if err == nil {
				go am.manager.handleNodeTimeout(addr)
				delete(am.lastRecv, addrStr)
				delete(am.lastSent, addrStr)
			}
		}
	}
}

func (am *ActivityManager) RecordMessageSent(addr *net.UDPAddr) {
	am.mu.Lock()
	defer am.mu.Unlock()

	addrStr := addr.String()
	am.lastSent[addrStr] = time.Now()
}

func (am *ActivityManager) RecordMessageReceived(addr *net.UDPAddr) {
	am.mu.Lock()
	defer am.mu.Unlock()
	addrStr := addr.String()
	am.lastRecv[addrStr] = time.Now()
	am.lastSent[addrStr] = time.Now()
}

func (am *ActivityManager) AddNodeToMonitor(addr *net.UDPAddr) {
	am.mu.Lock()
	defer am.mu.Unlock()
	addrStr := addr.String()
	now := time.Now()
	am.lastSent[addrStr] = now
	am.lastRecv[addrStr] = now
}

func (am *ActivityManager) RemoveNode(addr *net.UDPAddr) {
	am.mu.Lock()
	defer am.mu.Unlock()
	addrStr := addr.String()
	delete(am.lastSent, addrStr)
	delete(am.lastRecv, addrStr)
}

func (am *ActivityManager) Close() {
	if am.pingTicker != nil {
		am.pingTicker.Stop()
	}
	if am.timeoutTicker != nil {
		am.timeoutTicker.Stop()
	}
}
