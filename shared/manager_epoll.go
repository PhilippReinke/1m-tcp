package shared

import (
	"fmt"
	"net"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

// ConnManagerEpoll helps to keep track of connections. Methods are concurrency
// safe
type ConnManagerEpoll struct {
	mu    sync.Mutex
	fd    int
	conns map[int]*net.Conn
}

func NewConnManagerEpoll() *ConnManagerEpoll {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		panic(fmt.Sprintf("EpollCreate1 failed: %v", err))
	}
	return &ConnManagerEpoll{
		fd:    fd,
		conns: make(map[int]*net.Conn),
	}
}

func (m *ConnManagerEpoll) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.conns)
}

func (m *ConnManagerEpoll) AddConn(conn *net.Conn) {
	fd := getFD(conn)
	err := unix.EpollCtl(m.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		panic(fmt.Sprintf("EpollCtl failed: %v", err))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[fd] = conn
}

func (m *ConnManagerEpoll) RemoveConn(conn *net.Conn) {
	fd := getFD(conn)
	err := unix.EpollCtl(m.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		panic(fmt.Sprintf("EpollCtl failed: %v", err))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, fd)
}

func (m *ConnManagerEpoll) Wait() ([]*net.Conn, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(m.fd, events, 100)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var conns []*net.Conn
	for i := 0; i < n; i++ {
		conn := m.conns[int(events[i].Fd)]
		conns = append(conns, conn)
	}
	return conns, nil
}

func getFD(conn *net.Conn) int {
	rawConn, err := (*conn).(*net.TCPConn).SyscallConn()
	if err != nil {
		panic(fmt.Sprintf("SyscallConn: %v", err))
	}

	var fd int
	err = rawConn.Control(func(descriptor uintptr) {
		fd = int(descriptor)
	})
	if err != nil {
		panic(fmt.Sprintf("rawConn.Control: %v", err))
	}

	return fd
}
