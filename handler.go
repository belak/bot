package bot

import "github.com/sorcix/irc"

type Handler interface {
	HandleEvent(b *Bot, m *irc.Message)
}

type HandlerFunc func(b *Bot, m *irc.Message)

func (f HandlerFunc) HandleEvent(b *Bot, m *irc.Message) {
	f(b, m)
}
