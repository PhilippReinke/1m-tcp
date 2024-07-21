package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PhilippReinke/1m-tcp/shared"
	"github.com/alexflint/go-arg"
)

type args struct {
	Host     string `arg:"--host" default:"localhost" help:"port of target tcp server"`
	Port     string `arg:"-p,--port" default:"8080" help:"host of target tcp server"`
	Num      int    `arg:"-n,--num" default:"1" help:"number of desired tcp connections"`
	DummyMsg bool   `arg:"-m,--dummyMsg" help:"send dummy messages every 10 seconds"`
}

func (args) Description() string {
	return "This is a tcp client that lets you open an arbitrary number of connections."
}

func main() {
	var args args
	arg.MustParse(&args)

	// print info about local IP
	ip, err := shared.LocalIP()
	if err != nil {
		fmt.Println("Could not determine local IP of this client:", err)
	} else {
		fmt.Println("Local IP of this client:", ip)
	}

	// print info about settings
	fmt.Println("Target host:", args.Host)
	fmt.Println("Target port:", args.Port)
	fmt.Println("Num of expected clients:", args.Num)

	// connect
	addr := fmt.Sprintf("%v:%v", args.Host, args.Port)
	fmt.Printf("\nConnecting clients to %v...\n", addr)
	conns := shared.NewConnManager()

	for range args.Num {
		conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", args.Host, args.Port))
		if err != nil {
			fmt.Printf("Error dialing %v: %v\n", addr, err)
			continue
		}
		defer conn.Close()
		conns.AddConn(conn)
	}
	fmt.Printf("Successfully connected %v clients\n", conns.Len())

	// block
	fmt.Println("Blocking, press ctrl+c to close app...")
	if args.DummyMsg {
		for {
			conns.WriteDummyMsg()
			time.Sleep(time.Second * 10)
		}
	} else {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		<-done
	}
}
