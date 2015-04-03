package bot

import "log"

const (
	cmdPrefix = "!"

	botHost = "chat.freenode.net:6667"

	botNick = "bitbucket"
	botUser = "bitbucket"
	botName = "bitbucket"
	botPass = "@!sdalk1109jd"
)

func main() {
	b := NewBot(botNick, botUser, botName, botPass)
	err := b.Dial(botHost)
	if err != nil {
		log.Fatalln(err)
	}
}
