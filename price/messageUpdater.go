package price

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ad/anontrackerbot/config"
	"github.com/go-telegram/bot"
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

	data, err := GetData(dataURL)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (su *Price) Run(b *bot.Bot) error {
	updateTicker := time.NewTicker(time.Duration(su.config.UPDATE_DELAY) * time.Second)

	go func(b *bot.Bot) {
		for range updateTicker.C {
			data, err := su.Get()
			if err != nil {
				fmt.Println(err)
			}

			msg := Replacer(su.config.MESSAGE_FORMAT, data)

			for _, target := range su.config.UpdateMessageList {
				_, err := b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
					ChatID:    fmt.Sprintf("%d_%d", target.ChatID, target.MessageThreadID),
					MessageID: int(target.MessageID),
					Text:      msg,
				})

				if err != nil {
					fmt.Println(err)

					continue
				}

				// fmt.Println(m)

				// sender.MakeRequestDeferred(sndr.DeferredMessage{
				// 	Method: "sendMessage",
				// 	ChatID: target.ChatID,
				// 	Text:   msg,
				// }, sender.SendResult)
			}
		}
	}(b)

	return nil
}
