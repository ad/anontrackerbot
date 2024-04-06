package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/ad/anontrackerbot/config"
	"github.com/ad/anontrackerbot/logger"
	"github.com/ad/anontrackerbot/price"
	"github.com/ad/anontrackerbot/sender"
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

	sndr, errInitSender := sender.InitSender(ctx, lgr, conf, pr)
	if errInitSender != nil {
		return errInitSender
	}

	if len(conf.TelegramAdminIDsList) != 0 {
		sndr.MakeRequestDeferred(sender.DeferredMessage{
			Method: "sendMessage",
			ChatID: conf.TelegramAdminIDsList[0],
			Text:   "Bot restarted",
		}, sndr.SendResult)
	}

	pr.Run(sndr.Bot)

	return nil
}
