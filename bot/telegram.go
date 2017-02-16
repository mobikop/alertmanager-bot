package bot

import (
	"time"

	"github.com/tucnak/telebot"
)

// TelegramBroker implements the Broker interface and
// allows communication between the bot and telegram
type TelegramBroker struct {
	engine   *Engine
	telegram *telebot.Bot
}

// NewTelegramBroker returns a TelegramBroker that's connected to telegram
func NewTelegramBroker(e *Engine, Token string) (*TelegramBroker, error) {
	telegram, err := telebot.NewBot(Token)
	if err != nil {
		return nil, err
	}

	return &TelegramBroker{
		engine:   e,
		telegram: telegram,
	}, nil
}

// Name returns the name of the Broker: telegram
func (b *TelegramBroker) Name() string {
	return "telegram"
}

// Run the TelegramBroker and receive incoming messages via channel
func (b *TelegramBroker) Run(done chan<- bool, in chan<- Context) {
	messages := make(chan telebot.Message, 100)
	b.telegram.Listen(messages, time.Second)

	for message := range messages {
		b.telegram.SendChatAction(message.Chat, telebot.Typing)

		ctx := &TelegramContext{broker: b, message: message}

		if handlers, ok := b.engine.commands[message.Text]; ok {
			handlers1 := append(b.engine.middlewares, handlers...)
			for _, handler := range handlers1 {
				if err := handler(ctx); err != nil {
					b.telegram.SendMessage(message.Chat, err.Error(), nil)
					break
				}
			}
		} else {
			for _, handler := range b.engine.notFound {
				if err := handler(ctx); err != nil {
					b.telegram.SendMessage(message.Chat, err.Error(), nil)
					break
				}
			}
		}
	}

	done <- true
}

// TelegramContext implements the Context interface and
// makes sure everything is passed on to telegram
type TelegramContext struct {
	// TODO
	//ctx     context.Context
	broker  *TelegramBroker
	message telebot.Message
}

// Broker returns the name of the broker
func (c *TelegramContext) Broker() string {
	return c.broker.Name()
}

// Raw returns the raw text of the incoming message
func (c *TelegramContext) Raw() string {
	return c.message.Text
}

// User returns the user of the incoming message
func (c *TelegramContext) User() telebot.User {
	return c.message.Sender
}

// Write

// String sends a string back to the user
func (c *TelegramContext) String(msg string) error {
	return c.broker.telegram.SendMessage(c.message.Chat, msg, nil)
}

// Markdown sends a markdown formatted string back to the user
func (c *TelegramContext) Markdown(msg string) error {
	options := &telebot.SendOptions{ParseMode: telebot.ModeMarkdown}
	return c.broker.telegram.SendMessage(c.message.Chat, msg, options)
}