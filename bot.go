package bot

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/sorcix/irc"
)

func MessageFromChannel(m *irc.Message) bool {
	if len(m.Params) == 0 {
		return false
	}

	loc := m.Params[0]
	return len(loc) > 0 && (loc[0] == irc.Channel || loc[0] == irc.Distributed)
}

type Bot struct {
	Handler *BasicMux

	currentNick string
	user        string
	name        string
	pass        string

	// Internal things
	conn *irc.Conn
	err  error
}

func NewBot(nick, user, name, pass string) *Bot {
	b := &Bot{
		NewBasicMux(),
		nick,
		user,
		name,
		pass,
		nil,
		nil,
	}

	return b
}

func (b *Bot) Send(m *irc.Message) {
	if b.err != nil {
		return
	}

	err := b.conn.Encode(m)
	if err != nil {
		b.err = err
	}
}

func prepend(v []interface{}, e interface{}) []interface{} {
	var vc []interface{}

	vc = append(vc, e)
	vc = append(vc, v...)

	return vc
}

func (b *Bot) Reply(m *irc.Message, format string, v ...interface{}) {
	if len(m.Params) == 0 || len(m.Params[0]) == 0 {
		log.Println("Invalid IRC event")
		return
	}

	// Create the base message
	out := &irc.Message{
		Command: "PRIVMSG",
	}

	// Make sure we send it to the right place
	if MessageFromChannel(m) {
		out.Params = append(out.Params, m.Params[0])
	} else {
		out.Params = append(out.Params, m.Prefix.Name)
	}

	// Append the outgoing text
	out.Params = append(out.Params, fmt.Sprintf(format, v...))

	b.Send(out)
}

func (b *Bot) MentionReply(m *irc.Message, format string, v ...interface{}) {
	if len(m.Params) == 0 || len(m.Params[0]) == 0 {
		log.Println("Invalid IRC event")
		return
	}

	if MessageFromChannel(m) {
		format = "%s: " + format
		v = prepend(v, m.Prefix.Name)
	}

	b.Reply(m, format, v...)
}

func (b *Bot) mainLoop(conn io.ReadWriteCloser) error {
	b.conn = irc.NewConn(conn)

	// Startup commands
	if len(b.pass) > 0 {
		b.Send(&irc.Message{
			Command: "PASS",
			Params:  []string{b.pass},
		})
	}

	b.Send(&irc.Message{
		Command: "NICK",
		Params:  []string{b.currentNick},
	})

	b.Send(&irc.Message{
		Command: "USER",
		Params:  []string{b.user, "0.0.0.0", "0.0.0.0", b.name},
	})

	var m *irc.Message
	for {
		m, b.err = b.conn.Decode()
		if b.err != nil {
			break
		}

		if m.Command == "PING" {
			log.Println("Sending PONG")
			b.Send(&irc.Message{
				Command: "PONG",
				Params:  []string{m.Trailing()},
			})
		} else if m.Command == "PONG" {
			ns, _ := strconv.ParseInt(m.Trailing(), 10, 64)
			delta := time.Duration(time.Now().UnixNano() - ns)

			log.Println("!!! Lag:", delta)
		} else if m.Command == "NICK" {
			if m.Prefix.Name == b.currentNick && len(m.Params) > 0 {
				b.currentNick = m.Params[0]
			}
		} else if m.Command == "001" {
			if len(m.Params) > 0 {
				b.currentNick = m.Params[0]
			}
		} else if m.Command == "437" || m.Command == "433" {
			b.currentNick = b.currentNick + "_"
			b.Send(&irc.Message{
				Command: "NICK",
				Params:  []string{b.currentNick},
			})
		}

		log.Println(m)

		b.Handler.HandleEvent(b, m)

		// TODO: Make this work better
		if b.err != nil {
			break
		}
	}

	return b.err
}

func (b *Bot) DialTLS(host string, conf *tls.Config) error {
	tcpConn, err := tls.Dial("tcp", host, conf)
	if err != nil {
		return err
	}

	return b.mainLoop(tcpConn)
}

func (b *Bot) Dial(host string) error {
	tcpConn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}

	return b.mainLoop(tcpConn)
}
