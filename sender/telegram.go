package sender

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/ad/anontrackerbot/config"
	"github.com/ad/anontrackerbot/price"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Sender struct {
	sync.RWMutex
	logger           *slog.Logger
	config           *config.Config
	Bot              *bot.Bot
	Config           *config.Config
	Price            *price.Price
	deferredMessages map[int64]chan DeferredMessage
	lastMessageTimes map[int64]int64
}

func InitSender(ctx context.Context, logger *slog.Logger, config *config.Config, pr *price.Price) (*Sender, error) {
	sender := &Sender{
		logger:           logger,
		config:           config,
		Price:            pr,
		deferredMessages: make(map[int64]chan DeferredMessage),
		lastMessageTimes: make(map[int64]int64),
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(sender.handler),
		bot.WithSkipGetMe(),
	}

	b, newBotError := bot.New(config.TelegramToken, opts...)
	if newBotError != nil {
		return nil, fmt.Errorf("start bot error: %s", newBotError)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/price", bot.MatchTypePrefix, sender.replyWithCoin)

	go b.Start(ctx)
	go sender.sendDeferredMessages()

	sender.Bot = b

	return sender, nil
}

func (s *Sender) handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if s.config.Debug {
		s.logger.Debug(formatUpdateForLog(update))
	}
}

func (s *Sender) replyWithCoin(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !slices.Contains(s.config.TelegramAdminIDsList, update.Message.From.ID) {
		return
	}

	pr, err := s.Price.Get()
	if err != nil {
		fmt.Println(err)
	}

	// s.logger.Info(fmt.Sprintf("New message %d:%d:%d", update.Message.Chat.ID, update.Message.MessageThreadID, update.Message.ID))

	msg, err := b.SendMessage(
		context.Background(),
		&bot.SendMessageParams{
			ChatID:          fmt.Sprintf("%d_%d", update.Message.Chat.ID, update.Message.MessageThreadID),
			Text:            price.Format(pr),
			MessageThreadID: update.Message.MessageThreadID,
		},
	)

	if err != nil {
		fmt.Println(err)
	}

	if msg != nil {
		s.logger.Info(fmt.Sprintf("New message %d:%d:%d", msg.Chat.ID, msg.MessageThreadID, msg.ID))
	}
}
