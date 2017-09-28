package rpcx

import "time"

type TickArgs struct {
	DeliveredAt time.Time
	ReceivedAt  time.Time
}

type TickReply struct {
	ReceivedAt  time.Time
	DeliveredAt time.Time
}
