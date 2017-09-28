package rpcx

import (
	"log"
	"testing"
	"time"
)

type pingServer struct {
}

func (s *pingServer) PING(args *PingArgs, reply *PingReply) error {
	log.Printf("[PING SERVER] Received PING\r\n")
	return nil
}

func (s *pingServer) TICK(args *TickArgs, reply *TickReply) error {
	log.Printf("[TICK SERVER] Received TICK\r\n")
	return nil
}

func TestTickHandler(t *testing.T) {
	server := &Server{
		Delegate: &pingServer{},
	}

	server.RegisterInternalServices()
	go server.Run(":28889")

	client := &Client{
		Address: "127.0.0.1:28889",
	}

	client.TickHandler = func(c *Client, args *TickArgs, reply *TickReply, err error) {
		args.ReceivedAt = time.Now()
		duration := args.ReceivedAt.Sub(args.DeliveredAt)
		log.Printf("[TICK CLIENT] TICK Address=%s Duration=%v", c.Address, duration)

		if nil != err {
			t.Error(err)
		} else {
			t.Logf("[TICK CLIENT] TICK Address=%s Duration=%v", c.Address, duration)
		}

		server.Stop()
		c.StopTick()

	}

	client.StartTick(time.Second)
}

func TestPingHandler(t *testing.T) {
	server := &Server{
		Delegate: &pingServer{},
	}

	server.RegisterInternalServices()
	go server.Run(":28888")

	//defer server.Stop()

	client := &Client{
		Address: "127.0.0.1:28888",
	}

	client.PingHandler = func(c *Client, ping *PingArgs, pong *PingReply, err error) {
		ping.ReceivedAt = time.Now()
		duration := ping.ReceivedAt.Sub(ping.DeliveredAt)
		log.Printf("[PING CLIENT] PING Address=%s Duration=%v", c.Address, duration)

		if nil != err {
			t.Error(err)
		} else {
			t.Logf("[PING CLIENT] PING Address=%s Duration=%v", c.Address, duration)
		}

		server.Stop()
	}

	client.Ping()
}
