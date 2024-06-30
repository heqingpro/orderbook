package orderbook

import (
	"errors"
	"time"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/shopspring/decimal"
)

type AggregatedOrderBook struct {
	Exchange   string
	Symbol     string
	Timestamp  time.Time
	Aggregated *orderBook
	Local      *orderBook
}

type orderBook struct {
	Bids      *rbt.Tree
	Asks      *rbt.Tree
	BidsMap   map[decimal.Decimal]decimal.Decimal
	AsksMap   map[decimal.Decimal]decimal.Decimal
	Timestamp time.Time
}

func newOrderBook() *orderBook {
	return &orderBook{
		Bids:    rbt.NewWith(BidComparator),
		Asks:    rbt.NewWith(AskComparator),
		BidsMap: make(map[decimal.Decimal]decimal.Decimal, 0),
		AsksMap: make(map[decimal.Decimal]decimal.Decimal, 0),
	}
}

func NewFromLocalOrderBook(Symbol string, orders []LocalOrderUpdate) (*AggregatedOrderBook, error) {
	local := newOrderBook()
	err := local.updateLocalOrder(orders)
	if err != nil {
		return nil, err
	}
	return &AggregatedOrderBook{
		Symbol:    Symbol,
		Local:     local,
		Timestamp: time.Now(),
	}, nil
}

func (a *AggregatedOrderBook) UpdateSnapshot(s Snapshot) {
	a.Exchange = s.Exchange
	a.Timestamp = s.Timestamp
	aggregated := newOrderBook()
	aggregated.updateSnapshot(s)
	// Aggregate local data
	a.aggregateOrderBook()
}

func (a *AggregatedOrderBook) aggregateOrderBook() {
	for price, quantity := range a.Local.AsksMap {
		q, ok := a.Aggregated.AsksMap[price]
		if !ok {
			a.Aggregated.Asks.Put(price, nil)
			a.Aggregated.AsksMap[price] = quantity
		} else {
			a.Aggregated.AsksMap[price] = q.Add(quantity)
		}
	}
	for price, quantity := range a.Local.BidsMap {
		q, ok := a.Aggregated.BidsMap[price]
		if !ok {
			a.Aggregated.Bids.Put(price, nil)
			a.Aggregated.BidsMap[price] = quantity
		} else {
			a.Aggregated.BidsMap[price] = q.Add(quantity)
		}
	}
}

func (a *AggregatedOrderBook) UpdateQuote(s QuoteStream) {
	a.Timestamp = s.Timestamp
	for _, ask := range s.Asks {
		price, quantity := ask[0], ask[1]
		_, ok := a.Aggregated.AsksMap[price]
		if !ok {
			a.Aggregated.Asks.Put(price, nil)
			a.Aggregated.AsksMap[price] = quantity
		} else {
			quantityNew := quantity.Add(a.Local.AsksMap[price])
			if quantityNew.IsZero() {
				a.Aggregated.Asks.Remove(price)
				delete(a.Aggregated.AsksMap, price)
			} else {
				a.Aggregated.AsksMap[price] = quantityNew
			}
		}
	}
	for _, b := range s.Bids {
		price, quantity := b[0], b[1]
		_, ok := a.Aggregated.BidsMap[price]
		if !ok {
			a.Aggregated.Bids.Put(price, nil)
			a.Aggregated.BidsMap[price] = quantity
		} else {
			quantityNew := quantity.Add(a.Local.BidsMap[price])
			if quantityNew.IsZero() {
				a.Aggregated.Bids.Remove(price)
				delete(a.Aggregated.BidsMap, price)
			} else {
				a.Aggregated.BidsMap[price] = quantityNew
			}
		}
	}
}

func (a *AggregatedOrderBook) UpdateLocalOrder(orders []LocalOrderUpdate) error {
	if err := a.Local.updateLocalOrder(orders); err != nil {
		return err
	}
	if err := a.Aggregated.updateLocalOrder(orders); err != nil {
		return err
	}
	return nil
}

func (a *AggregatedOrderBook) GetOrderBook() Snapshot {
	s := Snapshot{
		Exchange:  a.Exchange,
		Symbol:    a.Symbol,
		Timestamp: a.Timestamp,
		Bids:      [][2]decimal.Decimal{},
		Asks:      [][2]decimal.Decimal{},
	}
	ait := a.Aggregated.Asks.Iterator()
	for ait.Next() {
		price := ait.Key().(decimal.Decimal)
		s.Asks = append(s.Asks, [2]decimal.Decimal{price, a.Aggregated.AsksMap[price]})
	}
	bit := a.Aggregated.Bids.Iterator()
	for bit.Next() {
		price := bit.Key().(decimal.Decimal)
		s.Bids = append(s.Bids, [2]decimal.Decimal{price, a.Aggregated.BidsMap[price]})
	}
	return s
}

func (o *orderBook) updateSnapshot(s Snapshot) {
	for _, b := range s.Bids {
		o.Bids.Put(b[0], nil)
		o.BidsMap[b[0]] = b[1]
	}
	for _, a := range s.Asks {
		o.Asks.Put(a[0], nil)
		o.AsksMap[a[0]] = a[1]
	}
}

func (o *orderBook) updateLocalOrder(orders []LocalOrderUpdate) error {
	var priceMap map[decimal.Decimal]decimal.Decimal
	var tree *rbt.Tree
	var quantityNew decimal.Decimal
	for _, order := range orders {
		switch order.Side {
		case Bid:
			priceMap = o.BidsMap
			tree = o.Bids
		case Ask:
			priceMap = o.AsksMap
			tree = o.Asks
		}
		quantity, ok := priceMap[order.Price]
		switch order.Operate {
		case Add:
			if ok {
				quantityNew = quantity.Add(order.Quantity)
			} else {
				quantityNew = order.Quantity
			}
		case Delete:
			if ok {
				quantityNew = quantity.Sub(order.Quantity)
			} else {
				return errors.New("local orders invalid")
			}
		}
		if quantityNew.IsZero() {
			tree.Remove(order.Price)
			delete(priceMap, order.Price)
		} else {
			tree.Put(order.Price, nil)
			priceMap[order.Price] = quantityNew
		}
		o.Timestamp = order.Timestamp
	}
	return nil
}

func AskComparator(a, b interface{}) int {
	aAsserted := a.(decimal.Decimal)
	bAsserted := b.(decimal.Decimal)
	switch {
	case aAsserted.GreaterThan(bAsserted):
		return 1
	case aAsserted.LessThan(bAsserted):
		return -1
	default:
		return 0
	}
}

func BidComparator(a, b interface{}) int {
	aAsserted := a.(decimal.Decimal)
	bAsserted := b.(decimal.Decimal)
	switch {
	case aAsserted.GreaterThan(bAsserted):
		return -1
	case aAsserted.LessThan(bAsserted):
		return 1
	default:
		return 0
	}
}
