package ticker

import (
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges/assets"
)

func TestPriceToString(t *testing.T) {
	newPair := pair.NewCurrencyPair("BTC", "USD")
	priceStruct := Price{
		Pair:         newPair,
		CurrencyPair: newPair.Pair().String(),
		Last:         1200,
		High:         1298,
		Low:          1148,
		Bid:          1195,
		Ask:          1220,
		Volume:       5,
		PriceATH:     1337,
	}

	newTicker := CreateNewTicker("ANX", newPair, priceStruct, assets.AssetTypeSpot)

	if newTicker.PriceToString(newPair, "last", assets.AssetTypeSpot) != "1200" {
		t.Error("Test Failed - ticker PriceToString last value is incorrect")
	}
	if newTicker.PriceToString(newPair, "high", assets.AssetTypeSpot) != "1298" {
		t.Error("Test Failed - ticker PriceToString high value is incorrect")
	}
	if newTicker.PriceToString(newPair, "low", assets.AssetTypeSpot) != "1148" {
		t.Error("Test Failed - ticker PriceToString low value is incorrect")
	}
	if newTicker.PriceToString(newPair, "bid", assets.AssetTypeSpot) != "1195" {
		t.Error("Test Failed - ticker PriceToString bid value is incorrect")
	}
	if newTicker.PriceToString(newPair, "ask", assets.AssetTypeSpot) != "1220" {
		t.Error("Test Failed - ticker PriceToString ask value is incorrect")
	}
	if newTicker.PriceToString(newPair, "volume", assets.AssetTypeSpot) != "5" {
		t.Error("Test Failed - ticker PriceToString volume value is incorrect")
	}
	if newTicker.PriceToString(newPair, "ath", assets.AssetTypeSpot) != "1337" {
		t.Error("Test Failed - ticker PriceToString ath value is incorrect")
	}
	if newTicker.PriceToString(newPair, "obtuse", assets.AssetTypeSpot) != "" {
		t.Error("Test Failed - ticker PriceToString obtuse value is incorrect")
	}
}

func TestGetTicker(t *testing.T) {
	newPair := pair.NewCurrencyPair("BTC", "USD")
	priceStruct := Price{
		Pair:         newPair,
		CurrencyPair: newPair.Pair().String(),
		Last:         1200,
		High:         1298,
		Low:          1148,
		Bid:          1195,
		Ask:          1220,
		Volume:       5,
		PriceATH:     1337,
	}

	ProcessTicker("bitfinex", newPair, priceStruct, assets.AssetTypeSpot)
	tickerPrice, err := GetTicker("bitfinex", newPair, assets.AssetTypeSpot)
	if err != nil {
		t.Errorf("Test Failed - Ticker GetTicker init error: %s", err)
	}
	if tickerPrice.CurrencyPair != "BTCUSD" {
		t.Error("Test Failed - ticker tickerPrice.CurrencyPair value is incorrect")
	}

	_, err = GetTicker("blah", newPair, assets.AssetTypeSpot)
	if err == nil {
		t.Fatal("Test Failed. TestGetTicker returned nil error on invalid exchange")
	}

	newPair.FirstCurrency = "ETH"
	_, err = GetTicker("bitfinex", newPair, assets.AssetTypeSpot)
	if err == nil {
		t.Fatal("Test Failed. TestGetTicker returned ticker for invalid first currency")
	}

	btcltcPair := pair.NewCurrencyPair("BTC", "LTC")
	_, err = GetTicker("bitfinex", btcltcPair, assets.AssetTypeSpot)
	if err == nil {
		t.Fatal("Test Failed. TestGetTicker returned ticker for invalid second currency")
	}

	priceStruct.PriceATH = 9001
	ProcessTicker("bitfinex", newPair, priceStruct, "futures_3m")
	tickerPrice, err = GetTicker("bitfinex", newPair, "futures_3m")
	if err != nil {
		t.Errorf("Test Failed - Ticker GetTicker init error: %s", err)
	}

	if tickerPrice.PriceATH != 9001 {
		t.Error("Test Failed - ticker tickerPrice.PriceATH value is incorrect")
	}
}

func TestGetTickerByExchange(t *testing.T) {
	newPair := pair.NewCurrencyPair("BTC", "USD")
	priceStruct := Price{
		Pair:         newPair,
		CurrencyPair: newPair.Pair().String(),
		Last:         1200,
		High:         1298,
		Low:          1148,
		Bid:          1195,
		Ask:          1220,
		Volume:       5,
		PriceATH:     1337,
	}

	anxTicker := CreateNewTicker("ANX", newPair, priceStruct, assets.AssetTypeSpot)
	Tickers = append(Tickers, anxTicker)

	tickerPtr, err := GetTickerByExchange("ANX")
	if err != nil {
		t.Errorf("Test Failed - GetTickerByExchange init error: %s", err)
	}
	if tickerPtr.ExchangeName != "ANX" {
		t.Error("Test Failed - GetTickerByExchange ExchangeName value is incorrect")
	}
}

func TestFirstCurrencyExists(t *testing.T) {
	newPair := pair.NewCurrencyPair("BTC", "USD")
	priceStruct := Price{
		Pair:         newPair,
		CurrencyPair: newPair.Pair().String(),
		Last:         1200,
		High:         1298,
		Low:          1148,
		Bid:          1195,
		Ask:          1220,
		Volume:       5,
		PriceATH:     1337,
	}

	alphaTicker := CreateNewTicker("alphapoint", newPair, priceStruct, assets.AssetTypeSpot)
	Tickers = append(Tickers, alphaTicker)

	if !FirstCurrencyExists("alphapoint", "BTC") {
		t.Error("Test Failed - FirstCurrencyExists1 value return is incorrect")
	}
	if FirstCurrencyExists("alphapoint", "CATS") {
		t.Error("Test Failed - FirstCurrencyExists2 value return is incorrect")
	}
}

func TestSecondCurrencyExists(t *testing.T) {
	t.Parallel()

	newPair := pair.NewCurrencyPair("BTC", "USD")
	priceStruct := Price{
		Pair:         newPair,
		CurrencyPair: newPair.Pair().String(),
		Last:         1200,
		High:         1298,
		Low:          1148,
		Bid:          1195,
		Ask:          1220,
		Volume:       5,
		PriceATH:     1337,
	}

	bitstampTicker := CreateNewTicker("bitstamp", newPair, priceStruct, assets.AssetTypeSpot)
	Tickers = append(Tickers, bitstampTicker)

	if !SecondCurrencyExists("bitstamp", newPair) {
		t.Error("Test Failed - SecondCurrencyExists1 value return is incorrect")
	}

	newPair.SecondCurrency = "DOGS"
	if SecondCurrencyExists("bitstamp", newPair) {
		t.Error("Test Failed - SecondCurrencyExists2 value return is incorrect")
	}
}

func TestCreateNewTicker(t *testing.T) {
	newPair := pair.NewCurrencyPair("BTC", "USD")
	priceStruct := Price{
		Pair:         newPair,
		CurrencyPair: newPair.Pair().String(),
		Last:         1200,
		High:         1298,
		Low:          1148,
		Bid:          1195,
		Ask:          1220,
		Volume:       5,
		PriceATH:     1337,
	}

	newTicker := CreateNewTicker("ANX", newPair, priceStruct, assets.AssetTypeSpot)
	const float64Type = "float64"

	if reflect.ValueOf(newTicker).NumField() != 2 {
		t.Error("Test Failed - ticker CreateNewTicker struct change/or updated")
	}
	if reflect.TypeOf(newTicker.ExchangeName).String() != "string" {
		t.Error("Test Failed - ticker CreateNewTicker.ExchangeName value is not a string")
	}
	if newTicker.ExchangeName != "ANX" {
		t.Error("Test Failed - ticker CreateNewTicker.ExchangeName value is not ANX")
	}

	if newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].Pair.Pair().String() != "BTCUSD" {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].Pair.Pair().String() value is not expected 'BTCUSD'")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].Ask).String() != float64Type {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].Ask value is not a float64")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].Bid).String() != float64Type {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].Bid value is not a float64")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].CurrencyPair).String() != "string" {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].CurrencyPair value is not a string")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].High).String() != float64Type {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].High value is not a float64")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].Last).String() != float64Type {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].Last value is not a float64")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].Low).String() != float64Type {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].Low value is not a float64")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].PriceATH).String() != float64Type {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].PriceATH value is not a float64")
	}
	if reflect.TypeOf(newTicker.Price["BTC"]["USD"][assets.AssetTypeSpot].Volume).String() != float64Type {
		t.Error("Test Failed - ticker newTicker.Price[BTC][USD].Volume value is not a float64")
	}
}

func TestProcessTicker(t *testing.T) { // non-appending function to tickers
	Tickers = []Ticker{}
	newPair := pair.NewCurrencyPair("BTC", "USD")
	priceStruct := Price{
		Pair:         newPair,
		CurrencyPair: newPair.Pair().String(),
		Last:         1200,
		High:         1298,
		Low:          1148,
		Bid:          1195,
		Ask:          1220,
		Volume:       5,
		PriceATH:     1337,
	}

	ProcessTicker("btcc", newPair, priceStruct, assets.AssetTypeSpot)

	result, err := GetTicker("btcc", newPair, assets.AssetTypeSpot)
	if err != nil {
		t.Fatal("Test failed. TestProcessTicker failed to create and return a new ticker")
	}

	if result.Pair.Pair() != newPair.Pair() {
		t.Fatal("Test failed. TestProcessTicker pair mismatch")
	}

	secondPair := pair.NewCurrencyPair("BTC", "AUD")
	priceStruct.Pair = secondPair
	ProcessTicker("btcc", secondPair, priceStruct, assets.AssetTypeSpot)

	result, err = GetTicker("btcc", secondPair, assets.AssetTypeSpot)
	if err != nil {
		t.Fatal("Test failed. TestProcessTicker failed to create and return a new ticker")
	}

	result, err = GetTicker("btcc", newPair, assets.AssetTypeSpot)
	if err != nil {
		t.Fatal("Test failed. TestProcessTicker failed to return an existing ticker")
	}

	type quick struct {
		Name string
		P    pair.CurrencyPair
		TP   Price
	}

	var testArray []quick

	_ = rand.NewSource(time.Now().Unix())

	var wg sync.WaitGroup
	var sm sync.Mutex

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			newName := "Exchange" + strconv.FormatInt(rand.Int63(), 10)
			newPairs := pair.NewCurrencyPair("BTC"+strconv.FormatInt(rand.Int63(), 10),
				"USD"+strconv.FormatInt(rand.Int63(), 10))

			tp := Price{
				Pair:         newPairs,
				CurrencyPair: newPairs.Pair().String(),
				Last:         rand.Float64(),
			}

			ProcessTicker(newName, newPairs, tp, assets.AssetTypeSpot)
			sm.Lock()
			testArray = append(testArray, quick{Name: newName, P: newPairs, TP: tp})
			sm.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()

	for _, test := range testArray {
		wg.Add(1)
		fatalErr := false
		go func(test quick) {
			result, err := GetTicker(test.Name, test.P, assets.AssetTypeSpot)
			if err != nil {
				fatalErr = true
				return
			}

			if result.Last != test.TP.Last {
				t.Error("Test failed. TestProcessTicker failed bad values")
			}

			wg.Done()
		}(test)

		if fatalErr {
			t.Fatal("Test failed. TestProcessTicker failed to retrieve new ticker")
		}
	}
	wg.Wait()

}
