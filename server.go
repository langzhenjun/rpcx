package rpcx

import (
	"log"
	"net"
	"net/rpc"
	"time"
)

type ServerDelegate interface {
	PING(args *PingArgs, reply *PingReply) error
	TICK(args *TickArgs, reply *TickReply) error
}

type ServerStatus string

const (
	StatusStopped = ServerStatus("stopped")
	StatusPaused  = ServerStatus("paused")
	StatusRunning = ServerStatus("running")
)

// Server rpc server
type Server struct {
	Delegate     ServerDelegate
	rpcServer    *rpc.Server
	status       ServerStatus
	statusSwitch chan ServerStatus
	listener     net.Listener
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) RPCServer() *rpc.Server {
	if nil == s.rpcServer {
		s.rpcServer = rpc.NewServer()
		s.statusSwitch = make(chan ServerStatus)
	}

	return s.rpcServer
}

func (s *Server) RegisterInternalServices() {
	server := &internalServer{
		Delegate: s.Delegate,
	}
	s.RegisterServices("RPCX", server)
}

func (s *Server) RegisterServices(name string, rcvr interface{}) {
	if "" == name {
		s.RPCServer().Register(rcvr)
	} else {
		s.RPCServer().RegisterName(name, rcvr)
	}
}

func (s *Server) ServeConn(conn net.Conn) {
	//log.Printf("[RPCX SERVER] New Client Connected: %v", conn.RemoteAddr())
	go s.RPCServer().ServeConn(conn)
}

func (s *Server) Stop() {
	if s.status == StatusStopped {
		return
	}

	//log.Println("[RPCX SERVER] will Stop")

	// this will bring server from accepting connection
	s.listener.Close()
	s.status = StatusStopped
}

func (s *Server) Pause() {
	if s.status == StatusPaused {
		return
	}

	//log.Println("[RPCX SERVER] will pause")

	s.statusSwitch <- StatusPaused
	s.status = StatusPaused
}

// Run run
func (s *Server) Run(address string) {
	if s.status == StatusRunning {
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	s.listener = listener

	//log.Printf("[RPCX SERVER] Running on %s\r\n", address)

	for {
		if s.status == StatusStopped {
			//log.Printf("[RPCX SERVER] %s stopped\r\n", address)
			return
		}

		if s.status == StatusPaused {
			//log.Printf("[RPCX SERVER] %s paused\r\n", address)
			status := <-s.statusSwitch
			if StatusRunning != status {
				continue
			}
		}

		conn, err := listener.Accept()
		if err != nil {
			//log.Printf("[RPCX] Listener Accept Error: %v", err)
			s.Stop()
			//s.Delegate.Disconnected()
			continue
		}

		s.ServeConn(conn)
	}
}

type internalServer struct {
	Delegate ServerDelegate
}

func (t *internalServer) PING(args *PingArgs, reply *PingReply) error {
	reply.ReceivedAt = time.Now()
	reply.DeliveredAt = time.Now()

	if nil != t.Delegate {
		return t.Delegate.PING(args, reply)
	}

	return nil
}

func (t *internalServer) TICK(args *TickArgs, reply *TickReply) error {
	reply.ReceivedAt = time.Now()
	reply.DeliveredAt = time.Now()

	if nil != t.Delegate {
		return t.Delegate.TICK(args, reply)
	}

	return nil
}
