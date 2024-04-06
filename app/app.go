package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/ad/anontrackerbot/config"
	"github.com/ad/anontrackerbot/logger"
	"github.com/ad/anontrackerbot/price"
	sndr "github.com/ad/anontrackerbot/sender"
	"github.com/go-telegram/bot"
)

func Run(ctx context.Context, w io.Writer, args []string) error {
	conf, errInitConfig := config.InitConfig(os.Args)
	if errInitConfig != nil {
		return errInitConfig
	}

	lgr := logger.InitLogger(conf.Debug)

	// Recovery
	defer func() {
		if p := recover(); p != nil {
			lgr.Error(fmt.Sprintf("panic recovered: %s; stack trace: %s", p, string(debug.Stack())))
		}
	}()

	pr, err := price.InitPrice(lgr, conf)
	if err != nil {
		return err
	}

	sender, errInitSender := sndr.InitSender(ctx, lgr, conf, pr)
	if errInitSender != nil {
		return errInitSender
	}

	if len(conf.TelegramAdminIDsList) != 0 {
		sender.MakeRequestDeferred(sndr.DeferredMessage{
			Method: "sendMessage",
			ChatID: conf.TelegramAdminIDsList[0],
			Text:   "Bot restarted",
		}, sender.SendResult)
	}

	updateTicker := time.NewTicker(time.Duration(conf.UPDATE_DELAY) * time.Second)

	go func() {
		for range updateTicker.C {
			data, err := pr.Get()
			if err != nil {
				fmt.Println(err)
			}

			msg := price.Format(data)

			for _, target := range conf.UpdateMessageList {
				_, err := sender.Bot.EditMessageText(context.Background(), &bot.EditMessageTextParams{
					ChatID:    fmt.Sprintf("%d_%d", target.ChatID, target.MessageThreadID),
					MessageID: int(target.MessageID),
					Text:      fmt.Sprintf("%s\n%s", msg, time.Now().Format("2006-01-02 15:04:05")),
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
	}()

	return nil
}
