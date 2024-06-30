package orderbook

import (
	"time"

	"github.com/shopspring/decimal"
)

type Snapshot struct {
	Exchange  string
	Symbol    string
	Timestamp time.Time
	Bids      [][2]decimal.Decimal
	Asks      [][2]decimal.Decimal
}

type QuoteStream struct {
	Exchange  string
	Symbol    string
	Timestamp time.Time
	Bids      [][2]decimal.Decimal
	Asks      [][2]decimal.Decimal
}

type Operate string

const (
	Add    Operate = "add"
	Delete Operate = "delete"
)

type Side string

const (
	Bid Side = "bid"
	Ask Side = "ask"
)

type LocalOrderUpdate struct {
	Timestamp time.Time
	Side      Side
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Operate   Operate
}
