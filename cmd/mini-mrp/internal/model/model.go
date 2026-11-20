package model

import "time"

type Server string

const (
	Mt4Live1 Server = "mt4_live1"
	Mt5Live1 Server = "mt5_live1"
)

type ClosedTrade struct {
	ServerId   Server
	Ticket     int
	Account    int
	Symbol     string
	Lot        float64
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  int64
	ClosePrice int64
	Profit     int64
}
