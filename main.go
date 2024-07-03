package main

import (
	"encoding/json"
	"orderbook/db"
	"orderbook/orderbook"
	"orderbook/redis"
	"os"
	"time"

	"github.com/shopspring/decimal"
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

func main() {
	snapshot, err := os.ReadFile("./snapshot.json")
	if err != nil {
		panic(err)
	}
	update, err := os.ReadFile("./update.json")
	if err != nil {
		panic(err)
	}
	snapshots := &Message{}
	updates := &Message{}
	if err := json.Unmarshal(snapshot, snapshots); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(update, updates); err != nil {
		panic(err)
	}

	b, err := NewBinance()
	if err != nil {
		panic(err)
	}
	s, err := snapshots.ToSnapshot()
	if err != nil {
		panic(err)
	}
	b.OrderBook["BTC_USDT"].UpdateSnapshot(s)
	as := b.OrderBook["BTC_USDT"].GetAggregatedSnapshot()
	data1, err := json.Marshal(as)
	if err != nil {
		panic(err)
	}
	u, err := updates.ToQuotes()
	if err != nil {
		panic(err)
	}
	b.OrderBook["BTC_USDT"].UpdateQuote(u)
	as = b.OrderBook["BTC_USDT"].GetAggregatedSnapshot()
	data2, err := json.Marshal(as)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("./snapshot1.json", data1, 0644); err != nil {
		panic(err)
	}
	if err := os.WriteFile("./snapshot2.json", data2, 0644); err != nil {
		panic(err)
	}
}

type Message struct {
	Action string
	Data   []Data
}

type Data struct {
	Asks [][4]string
	Bids [][4]string
}

func (m *Message) ToSnapshot() (orderbook.Snapshot, error) {
	bids, asks, err := m.getQuotes()
	if err != nil {
		return orderbook.Snapshot{}, err
	}
	snapshot := orderbook.Snapshot{
		Timestamp: time.Now(),
		Bids:      bids,
		Asks:      asks,
	}
	return snapshot, nil
}

func (m *Message) ToQuotes() (orderbook.QuoteStream, error) {
	bids, asks, err := m.getQuotes()
	if err != nil {
		return orderbook.QuoteStream{}, err
	}
	quotes := orderbook.QuoteStream{
		Timestamp: time.Now(),
		Bids:      bids,
		Asks:      asks,
	}
	return quotes, nil
}

func (m *Message) getQuotes() ([][2]decimal.Decimal, [][2]decimal.Decimal, error) {
	bids := make([][2]decimal.Decimal, 0)
	asks := make([][2]decimal.Decimal, 0)
	for _, data := range m.Data {
		for _, bid := range data.Bids {
			price, err := decimal.NewFromString(bid[0])
			if err != nil {
				return nil, nil, err
			}
			amount, err := decimal.NewFromString(bid[1])
			if err != nil {
				return nil, nil, err
			}
			bids = append(bids, [2]decimal.Decimal{price, amount})
		}
		for _, ask := range data.Asks {
			price, err := decimal.NewFromString(ask[0])
			if err != nil {
				return nil, nil, err
			}
			amount, err := decimal.NewFromString(ask[1])
			if err != nil {
				return nil, nil, err
			}
			asks = append(asks, [2]decimal.Decimal{price, amount})
		}
	}
	return bids, asks, nil
}
