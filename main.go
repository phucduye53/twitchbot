package main

import (
	"time"
	"twitchbot/twitchbot"
)

func main() {
	myBot := twitchbot.TwitchBot{
		Channel:     "phucduye53",
		MsgRate:     time.Duration(20/30) * time.Millisecond,
		Name:        "phucduye53",
		Port:        "6667",
		PrivatePath: "./private/oauth.json",
		Server:      "irc.chat.twitch.tv",
	}
	myBot.Start()
}
