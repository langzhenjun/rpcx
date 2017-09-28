package rpcx

import (
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type ErrorHandler func(client *Client, args interface{}, err error)
type PingHandler func(client *Client, ping *PingArgs, pong *PingReply, err error)
type TickHandler func(client *Client, args *TickArgs, reply *TickReply, err error)

//Client SRPC Client
type Client struct {
	rpcClient      *rpc.Client
	connectionLock sync.Mutex

	Address          string
	KeepAlive        bool
	ReconnectOnError bool

	ticker *time.Ticker

	ErrorHandler ErrorHandler
	PingHandler  PingHandler
	TickHandler  TickHandler
}

func (c *Client) Reconnect() error {
	c.rpcClient = nil
	err := c.Connect()
	return err
}

// Connect connect to server
func (c *Client) Connect() error {
	if nil != c.rpcClient {
		return nil
	}

	c.connectionLock.Lock()
	defer c.connectionLock.Unlock()

	if nil == c.rpcClient {
		client, err := net.Dial("tcp", c.Address)
		if err != nil {
			return ErrOnRefused(c.Address)
		}

		//if c.KeepAlive {
		//	rpcClient.(*net.TCPConn).SetKeepAlive(true)
		//	rpcClient.(*net.TCPConn).SetKeepAlivePeriod(time.Minute)
		//} else {
		//	rpcClient.(*net.TCPConn).SetKeepAlive(false)
		//}

		c.rpcClient = rpc.NewClient(client)
	}

	return nil
}

// Call rpc client call the remote,
func (c *Client) Call(service string, args interface{}, reply interface{}) error {
	if nil == args || nil == reply {
		return ErrOnNullParams()
	}

	if nil == c.rpcClient {
		err := c.Connect()
		if nil != err {
			return err
		}
	}

	errOnCall := c.rpcClient.Call(service, args, reply)

	//log.Printf("[RPCX] Address=%s Service=%s Args=%v Reply=%v Duration=%v", c.Address, service, args, reply, duration)

	return errOnCall
}

// PingArgs ping server
func (c *Client) Ping() {
	var ping PingArgs
	var pong PingReply

	ping.DeliveredAt = time.Now()

	err := c.Call("RPCX.PING", &ping, &pong)

	if nil != c.PingHandler {
		c.PingHandler(c, &ping, &pong, err)
		return
	}

	if nil != err && nil != c.ErrorHandler {
		c.ErrorHandler(c, &ping, err)
	}
}

func (c *Client) StartTick(d time.Duration) {
	c.StopTick()
	c.ticker = time.NewTicker(d)

	select {
	case <-c.ticker.C:
		args := TickArgs{
			DeliveredAt: time.Now(),
		}
		reply := TickReply{}
		err := c.Call("RPCX.TICK", &args, &reply)
		if nil != err {
			if nil != c.ErrorHandler {
				c.ErrorHandler(c, args, err)
			} else {
				log.Printf("[RPCX TICK ERROR] %s", err.Error())
			}

			return
		}

		if nil != c.TickHandler {
			c.TickHandler(c, &args, &reply, err)
		}
	}
}

func (c *Client) StopTick() {
	if nil != c.ticker {
		c.ticker.Stop()
		c.ticker = nil
	}
}
