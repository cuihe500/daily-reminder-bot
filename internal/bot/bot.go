package bot

import (
	"time"

	tele "gopkg.in/telebot.v3"
)

// Bot represents the Telegram bot
type Bot struct {
	*tele.Bot
}

// NewBot creates a new Bot instance
func NewBot(token, apiEndpoint string) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	// Set custom API endpoint if provided
	if apiEndpoint != "" {
		pref.URL = apiEndpoint
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	return &Bot{Bot: b}, nil
}

// Start starts the bot
func (b *Bot) Start() {
	b.Bot.Start()
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.Bot.Stop()
}
