package db

import (
	"time"

	"github.com/shopspring/decimal"
)

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type State string

const (
	Open      State = "open"
	Cancelled State = "cancelled"
	Completed State = "completed"
)

type Order struct {
	ID        string
	Symbol    string
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Side      Side
	Timestamp time.Time
	State     State
}
type Repo interface {
	GetAvailableOrders(Symbol string) ([]Order, error)
}
