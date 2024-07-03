package exchange

import (
	"orderbook/db"
	"orderbook/orderbook"
	"orderbook/redis"
)

type Binance struct {
	OrderBook map[string]orderbook.Repo
	Redis     redis.Repo
	db        db.Repo
}

func NewBinance() (*Binance, error) {
	b := &Binance{OrderBook: make(map[string]orderbook.Repo, 0)}
	if err := b.initOrderBook("BTC_USDT"); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Binance) initOrderBook(symbol string) error {
	orderbook, err := orderbook.NewFromLocalOrderBook(symbol, []orderbook.LocalOrderUpdate{})
	if err != nil {
		return err
	}
	b.OrderBook[symbol] = orderbook
	return nil
}
