package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// BinanceKline represents the structure of a Binance API kline/candlestick response
type BinanceKline []interface{}

// Fetch retrieves OHLCV data from Binance API
func Fetch(num, totime int64) ([]OHLCV, error) {
	url := "https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1m"
	req_url := url + "&limit=" + strconv.FormatInt(num, 10) + "&endTime=" + strconv.FormatInt(totime, 10)

	// Log the request time in both UTC and local time for debugging
	utcTime := time.Unix(totime, 0).UTC()
	localTime := time.Unix(totime, 0)
	fmt.Printf("Fetching data - UTC: %s, Local: %s\n",
		utcTime.Format("2006-01-02 15:04:05"),
		localTime.Format("2006-01-02 15:04:05"))
	fmt.Println("Request URL:", req_url)

	resp, err := http.Get(req_url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("http code " + strconv.Itoa(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseBinanceResponse(body)
}

// parseBinanceResponse converts Binance API response to our OHLCV format
func parseBinanceResponse(body []byte) ([]OHLCV, error) {
	var klines []BinanceKline
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, err
	}

	var ohlcvData []OHLCV
	for _, k := range klines {
		if len(k) < 11 {
			continue // Skip malformed entries
		}

		openTime := int64(k[0].(float64)) // Keep as milliseconds
		open, _ := strconv.ParseFloat(k[1].(string), 64)
		high, _ := strconv.ParseFloat(k[2].(string), 64)
		low, _ := strconv.ParseFloat(k[3].(string), 64)
		closePrice, _ := strconv.ParseFloat(k[4].(string), 64)
		volume, _ := strconv.ParseFloat(k[5].(string), 64)

		ohlcvData = append(ohlcvData, OHLCV{
			Time:   openTime,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		})
	}

	return ohlcvData, nil
}
