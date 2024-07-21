package main

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/PhilippReinke/1m-tcp/shared"
)

func main() {
	// increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Println("syscall.Getrlimit failed:", err)
		os.Exit(1)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Println("syscall.Setrlimit failed:", err)
		os.Exit(1)
	}

	// print info about local IP
	ip, err := shared.LocalIP()
	if err != nil {
		fmt.Println("Could not determine local IP:", err)
		os.Exit(1)
	} else {
		fmt.Println("Local IP of this server:", ip)
	}

	// print info about settings
	addr := fmt.Sprintf("%v:8080", ip)
	fmt.Println("Listing addr:", addr)

	// listen
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Could not listen on %v: %v\n", addr, err)
		os.Exit(1)
	}
	defer l.Close()

	conns := shared.NewConnManagerEpoll()
	go acceptConns(l, conns)
	go readLoopEpoll(conns)

	for {
		fmt.Println("num of conns:", conns.Len())
		time.Sleep(time.Second * 2)
	}
}

func acceptConns(l net.Listener, conns *shared.ConnManagerEpoll) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		conns.AddConn(&conn)
	}
}

func readLoopEpoll(conns *shared.ConnManagerEpoll) {
	for {
		connsEpoll, err := conns.Wait()
		if err != nil {
			// fmt.Println("conns.Wait() failed:", err)
			continue
		}

		buffer := make([]byte, 512)
		for _, conn := range connsEpoll {
			_, err := (*conn).Read(buffer)
			if err != nil {
				conns.RemoveConn(conn)
				break
			}

			// fmt.Printf(string(buffer[:n]))
		}
	}
}
