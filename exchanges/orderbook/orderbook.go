package orderbook

import (
	"errors"
	"sync"
	"time"

	"github.com/thrasher-/gocryptotrader/currency/pair"
)

// Const values for orderbook package
const (
	ErrOrderbookForExchangeNotFound = "ticker for exchange does not exist"
	ErrPrimaryCurrencyNotFound      = "primary currency for orderbook not found"
	ErrSecondaryCurrencyNotFound    = "secondary currency for orderbook not found"

	Spot = "SPOT"
)

// Vars for the orderbook package
var (
	Orderbooks []Orderbook
	m          sync.Mutex
)

// Item stores the amount and price values
type Item struct {
	Amount float64
	Price  float64
	ID     int64
}

// Base holds the fields for the orderbook base
type Base struct {
	Pair         pair.CurrencyPair `json:"pair"`
	CurrencyPair string            `json:"CurrencyPair"`
	Bids         []Item            `json:"bids"`
	Asks         []Item            `json:"asks"`
	LastUpdated  time.Time         `json:"last_updated"`
	AssetType    string
}

// Orderbook holds the orderbook information for a currency pair and type
type Orderbook struct {
	Orderbook    map[pair.CurrencyItem]map[pair.CurrencyItem]map[string]Base
	ExchangeName string
}

// CalculateTotalBids returns the total amount of bids and the total orderbook
// bids value
func (o *Base) CalculateTotalBids() (amountCollated, total float64) {
	for _, x := range o.Bids {
		amountCollated += x.Amount
		total += x.Amount * x.Price
	}
	return amountCollated, total
}

// CalculateTotalAsks returns the total amount of asks and the total orderbook
// asks value
func (o *Base) CalculateTotalAsks() (amountCollated, total float64) {
	for _, x := range o.Asks {
		amountCollated += x.Amount
		total += x.Amount * x.Price
	}
	return amountCollated, total
}

// Update updates the bids and asks
func (o *Base) Update(bids, asks []Item) {
	o.Bids = bids
	o.Asks = asks
	o.LastUpdated = time.Now()
}

// GetOrderbook checks and returns the orderbook given an exchange name and
// currency pair if it exists
func GetOrderbook(exchange string, p pair.CurrencyPair, orderbookType string) (Base, error) {
	orderbook, err := GetOrderbookByExchange(exchange)
	if err != nil {
		return Base{}, err
	}

	if !FirstCurrencyExists(exchange, p.FirstCurrency) {
		return Base{}, errors.New(ErrPrimaryCurrencyNotFound)
	}

	if !SecondCurrencyExists(exchange, p) {
		return Base{}, errors.New(ErrSecondaryCurrencyNotFound)
	}

	return orderbook.Orderbook[p.FirstCurrency][p.SecondCurrency][orderbookType], nil
}

// GetOrderbookByExchange returns an exchange orderbook
func GetOrderbookByExchange(exchange string) (*Orderbook, error) {
	m.Lock()
	defer m.Unlock()
	for x := range Orderbooks {
		if Orderbooks[x].ExchangeName == exchange {
			return &Orderbooks[x], nil
		}
	}
	return nil, errors.New(ErrOrderbookForExchangeNotFound)
}

// FirstCurrencyExists checks to see if the first currency of the orderbook map
// exists
func FirstCurrencyExists(exchange string, currency pair.CurrencyItem) bool {
	m.Lock()
	defer m.Unlock()
	for _, y := range Orderbooks {
		if y.ExchangeName == exchange {
			if _, ok := y.Orderbook[currency]; ok {
				return true
			}
		}
	}
	return false
}

// SecondCurrencyExists checks to see if the second currency of the orderbook
// map exists
func SecondCurrencyExists(exchange string, p pair.CurrencyPair) bool {
	m.Lock()
	defer m.Unlock()
	for _, y := range Orderbooks {
		if y.ExchangeName == exchange {
			if _, ok := y.Orderbook[p.FirstCurrency]; ok {
				if _, ok := y.Orderbook[p.FirstCurrency][p.SecondCurrency]; ok {
					return true
				}
			}
		}
	}
	return false
}

// CreateNewOrderbook creates a new orderbook
func CreateNewOrderbook(exchangeName string, p pair.CurrencyPair, orderbookNew Base, orderbookType string) Orderbook {
	m.Lock()
	defer m.Unlock()
	orderbook := Orderbook{}
	orderbook.ExchangeName = exchangeName
	orderbook.Orderbook = make(map[pair.CurrencyItem]map[pair.CurrencyItem]map[string]Base)
	a := make(map[pair.CurrencyItem]map[string]Base)
	b := make(map[string]Base)
	b[orderbookType] = orderbookNew
	a[p.SecondCurrency] = b
	orderbook.Orderbook[p.FirstCurrency] = a
	Orderbooks = append(Orderbooks, orderbook)
	return orderbook
}

// ProcessOrderbook processes incoming orderbooks, creating or updating the
// Orderbook list
func ProcessOrderbook(exchangeName string, p pair.CurrencyPair, orderbookNew Base, orderbookType string) {
	if orderbookNew.Pair.Pair() == "" {
		// set Pair if not set
		orderbookNew.Pair = p
	}
	orderbookNew.CurrencyPair = p.Pair().String()
	orderbookNew.LastUpdated = time.Now()

	orderbook, err := GetOrderbookByExchange(exchangeName)
	if err != nil {
		CreateNewOrderbook(exchangeName, p, orderbookNew, orderbookType)
		return
	}

	if FirstCurrencyExists(exchangeName, p.FirstCurrency) {
		m.Lock()
		a := make(map[string]Base)
		a[orderbookType] = orderbookNew
		orderbook.Orderbook[p.FirstCurrency][p.SecondCurrency] = a
		m.Unlock()
		return
	}

	m.Lock()
	a := make(map[pair.CurrencyItem]map[string]Base)
	b := make(map[string]Base)
	b[orderbookType] = orderbookNew
	a[p.SecondCurrency] = b
	orderbook.Orderbook[p.FirstCurrency] = a
	m.Unlock()
}
