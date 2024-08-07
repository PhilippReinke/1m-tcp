package shared

import (
	"net"
	"sync"
)

// ConnManager helps to keep track of connections. Methods are concurrency safe
type ConnManager struct {
	mu    sync.Mutex
	conns map[net.Conn]byte
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		conns: make(map[net.Conn]byte),
	}
}

func (m *ConnManager) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.conns)
}

func (m *ConnManager) AddConn(conn net.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[conn] = byte(0)
}

func (m *ConnManager) RemoveConn(conn net.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, conn)
}

func (m *ConnManager) WriteDummyMsg() {
	m.mu.Lock()
	conns := make([]net.Conn, 0, len(m.conns))
	for conn := range m.conns {
		conns = append(conns, conn)
	}
	m.mu.Unlock()

	for _, conn := range conns {
		_, err := conn.Write([]byte("hello tcp"))
		if err != nil {
			m.RemoveConn(conn)
			continue
		}
	}
}
