package exchange

import (
	"github.com/samber/lo"
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
	orders, err := b.db.GetAvailableOrders(symbol)
	if err != nil {
		return err
	}
	orderbook, err := orderbook.NewFromLocalOrderBook(symbol, lo.Map(orders, func(order db.Order, _ int) orderbook.LocalOrderUpdate {
		side := orderbook.Bid
		if order.Side == db.Sell {
			side = orderbook.Ask
		}
		return orderbook.LocalOrderUpdate{
			Side:     side,
			Price:    order.Price,
			Quantity: order.Quantity,
			Operate:  orderbook.Add,
		}
	}))
	if err != nil {
		return err
	}
	b.OrderBook[symbol] = orderbook
	return nil
}
