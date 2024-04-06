package price

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/ad/anontrackerbot/config"
)

type Price struct {
	logger *slog.Logger
	config *config.Config
}

func InitPrice(logger *slog.Logger, config *config.Config) (*Price, error) {
	price := &Price{
		logger: logger,
		config: config,
	}

	return price, nil
}

func (su *Price) Get() (GeckoterminalResponse, error) {
	dataURL := su.config.DATA_URL

	data, err := getData(dataURL)
	if err != nil {
		return data, err
	}

	return data, nil
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

	if m5PricePercentage > 10 {
		emoji = "🚀"
	} else if m5PricePercentage > 50 {
		emoji = "🚀🚀🚀"
	} else if m5PricePercentage > 0 {
		emoji = "🟢"
	} else if m5PricePercentage < 0 {
		emoji = "🔴"
	}

	return fmt.Sprintf("%s ANON: %s 24H: %s MC: %s", emoji, humanizeSmallMoney(coinPrice), humanizeBigMoney(volume), humanizeBigMoney(mcap))
}

func humanizeBigMoney(money float64) string {
	if money < 1000 {
		return fmt.Sprintf("$%.2f", money)
	}

	if money < 1000000 {
		return fmt.Sprintf("$%.2fK", money/1000)
	}

	if money < 1000000000 {
		return fmt.Sprintf("$%.2fM", money/1000000)
	}

	return fmt.Sprintf("$%.2fB", money/1000000000)
}

func humanizeSmallMoney(money float64) string {
	if money > 0.01 {
		return fmt.Sprintf("$%.3f", money)
	}
	if money > 0.001 {
		return fmt.Sprintf("$%.4f", money)
	}
	if money > 0.0001 {
		return fmt.Sprintf("$%.5f", money)
	}

	return fmt.Sprintf("$%.2f", money)
}
