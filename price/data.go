package price

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type GeckoterminalResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			BaseTokenPriceUsd             string      `json:"base_token_price_usd"`
			BaseTokenPriceNativeCurrency  string      `json:"base_token_price_native_currency"`
			QuoteTokenPriceUsd            string      `json:"quote_token_price_usd"`
			QuoteTokenPriceNativeCurrency string      `json:"quote_token_price_native_currency"`
			BaseTokenPriceQuoteToken      string      `json:"base_token_price_quote_token"`
			QuoteTokenPriceBaseToken      string      `json:"quote_token_price_base_token"`
			Address                       string      `json:"address"`
			Name                          string      `json:"name"`
			PoolCreatedAt                 time.Time   `json:"pool_created_at"`
			FdvUsd                        string      `json:"fdv_usd"`
			MarketCapUsd                  interface{} `json:"market_cap_usd"`
			PriceChangePercentage         struct {
				M5  string `json:"m5"`
				H1  string `json:"h1"`
				H6  string `json:"h6"`
				H24 string `json:"h24"`
			} `json:"price_change_percentage"`
			Transactions struct {
				M5 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"m5"`
				M15 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"m15"`
				M30 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"m30"`
				H1 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"h1"`
				H24 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"h24"`
			} `json:"transactions"`
			VolumeUsd struct {
				M5  string `json:"m5"`
				H1  string `json:"h1"`
				H6  string `json:"h6"`
				H24 string `json:"h24"`
			} `json:"volume_usd"`
			ReserveInUsd string `json:"reserve_in_usd"`
		} `json:"attributes"`
		Relationships struct {
			BaseToken struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"base_token"`
			QuoteToken struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"quote_token"`
			Dex struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"dex"`
		} `json:"relationships"`
	} `json:"data"`
}

func GetData(dataURL string) (GeckoterminalResponse, error) {
	var data GeckoterminalResponse

	err := GetJson(dataURL, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func GetJson(dataURL string, target interface{}) error {
	req, err := http.NewRequest("GET", dataURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		return err
	}

	return nil
}

func Format(data GeckoterminalResponse) string {
	coinPrice, err := strconv.ParseFloat(data.Data.Attributes.BaseTokenPriceUsd, 64)
	if err != nil {
		coinPrice = 0.0
	}

	volume, err := strconv.ParseFloat(data.Data.Attributes.VolumeUsd.H24, 64)
	if err != nil {
		volume = 0.0
	}

	mcap, err := strconv.ParseFloat(data.Data.Attributes.FdvUsd, 64)
	if err != nil {
		mcap = 0.0
	}

	m5PricePercentage, err := strconv.ParseFloat(data.Data.Attributes.PriceChangePercentage.M5, 64)
	if err != nil {
		m5PricePercentage = 0
	}

	emoji := "🎱"

	if m5PricePercentage > 15 {
		emoji = "🚀🚀🚀"
	} else if m5PricePercentage > 5 {
		emoji = "🚀"
	} else if m5PricePercentage > 0 {
		emoji = "🟢"
	} else if m5PricePercentage < 0 {
		emoji = "🔴"
	}

	return fmt.Sprintf("%s ANON: %s 24H: %s MC: %s", emoji, humanizeMoney(coinPrice), humanizeMoney(volume), humanizeMoney(mcap))
}

func humanizeMoney(money float64) string {
	if money < 1 && money > 0.01 {
		return fmt.Sprintf("$%.3f", money)
	}
	if money < 1 && money > 0.001 {
		return fmt.Sprintf("$%.4f", money)
	}
	if money < 1 && money > 0.0001 {
		return fmt.Sprintf("$%.5f", money)
	}

	if money < 1000 {
		return fmt.Sprintf("$%.2f", money)
	}

	if money < 1000000 {
		return fmt.Sprintf("$%.2fK", money/1000)
	}

	if money < 1000000000 {
		return fmt.Sprintf("$%.2fM", money/1000000)
	}

	return fmt.Sprintf("$%.2f", money)
}

func Replacer(template string, data GeckoterminalResponse) string {
	wrapped := gabs.Wrap(data)
	jsonOutput := wrapped.String()
	wrapped, _ = gabs.ParseJSON([]byte(jsonOutput))

	re := regexp.MustCompile(`[F|S|E]{[^{}]*}`)

	substitutor := func(match string) string {
		isFloat := strings.HasPrefix(match, "F")
		isString := strings.HasPrefix(match, "S")
		isEmoji := strings.HasPrefix(match, "E")

		varName := match[2 : len(match)-1]

		value := ""
		if wrapped.ExistsP(varName) {
			switch {
			case isEmoji:
				tempVal := 0.00
				val, found := wrapped.Path(varName).Data().(float64)
				if found {
					tempVal = val
				} else {
					value = wrapped.Path(varName).Data().(string)
					val, err := strconv.ParseFloat(value, 64)
					if err == nil {
						tempVal = val
					}
				}

				if tempVal > 15 {
					value = "🚀🚀🚀"
				} else if tempVal > 5 {
					value = "🚀"
				} else if tempVal > 0 {
					value = "🟢"
				} else if tempVal < 0 {
					value = "🔴"
				} else {
					value = "🎱"
				}
			case isFloat:
				val, found := wrapped.Path(varName).Data().(float64)
				if found {
					value = humanizeMoney(val)
				} else {
					value = wrapped.Path(varName).Data().(string)
					val, err := strconv.ParseFloat(value, 64)
					if err == nil {
						value = humanizeMoney(val)
					}
				}
			case isString:
				value = wrapped.Path(varName).Data().(string)
			default:
				value = fmt.Sprintf("%v", wrapped.Path(varName).Data())
			}
		}

		return value
	}

	return re.ReplaceAllStringFunc(template, substitutor)
}
