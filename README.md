# bot

bot is a simple wrapper around https://github.com/sorcix/irc, designed to make
writing bots easier. It is based on https://github.com/belak/irc but with
message parsing now abstracted out.

## Muxes

Muxes are one of the most important parts. They make it possible to preprocess
messages and make them easier to handle. Currently, the only Mux included with
the base package is the BasicMux which allows you to register handlers to only
operate on messages with a certain command.

## Example

```go
package main

import (
	"log"

	"github.com/belak/bot"
	"github.com/sorcix/irc"
)

const (
	botHost = "chat.freenode.net:6697"

	botNick = "testbot"
	botUser = "bot"
	botName = "Herbert"
	botPass = ""
)

func main() {
	b := bot.NewBot(botNick, botUser, botName, botPass)

	// 001 is a welcome event, so we join channels there
	b.Handler.Event("001", func(b *bot.Bot, m *irc.Message) {
		b.Send(&irc.Message{
			Command: "JOIN",
			Params: []string{"#bot-test-chan"},
		})
	})

	// Echo replies back to everyone
	b.Handler.Event("PRIVMSG", func(b *bot.Bot, m *irc.Message) {
		b.MentionReply(m, "%s", m.Trailing())
	})

	err := b.DialTLS(botHost, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
```
