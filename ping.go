package rpcx

import (
	"time"
)

// PingArgs input
type PingArgs struct {
	DeliveredAt time.Time
	ReceivedAt  time.Time
}

// PingReply output
type PingReply struct {
	ReceivedAt  time.Time
	DeliveredAt time.Time
}
