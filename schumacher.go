/*
 *  schumacher, the IRC bot of the #formula1 channel at Quakenet.
 *  Copyright (C) 2021-2022  Vasco Costa (gluon)
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/thoj/go-ircevent"
	"strings"
)

var nick = "Schumacher_"     // Nick to be used by the bot.
var channels = "#motorsport" // Names of the channels to join.
var adminNick = "gluon"      // Nick used by the admin of the bot.
var poll bool
var quiz bool
var activeChannel string

func main() {
	flag.StringVar(&nick, "nick", "Schumacher_", "Nick to be used by the bot.")
	flag.StringVar(&channels, "channels", "#motorsport", "Names of the channels to join.")
	flag.Parse()
	c := make(chan [2]string)
	irccon := irc.IRC(nick, nick)
	irccon.AddCallback("001", func(event *irc.Event) {
		irccon.Join(channels)
	})
	irccon.AddCallback("366", func(event *irc.Event) {})
	err := irccon.Connect(server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		m := event.Message()
		command, err := parseCommand(m, event.Nick, event.Arguments[0])
		if err != nil {
			if poll || quiz {
				if activeChannel == event.Arguments[0] {
					c <- [2]string{event.Nick, event.Message()}
				}
			} else if strings.Contains(strings.ToLower(event.Message()), "http") {
				go tskHTMLTitle(irccon, event.Arguments[0], event.Message())
			}
		} else {
			switch strings.ToLower(command.Name) {
			case "a", "ask":
				cmdAsk(irccon, command.Channel, command.Args)
			case "c", "h", "commands", "help":
				cmdHelp(irccon, command.Channel, strings.Join(command.Args, ""))
			case "n", "next":
				cmdNext(irccon, command.Channel, command.Nick, strings.Join(command.Args, " "))
			case "ny", "notify":
				cmdNotify(irccon, command.Channel, command.Nick, command.Args)
			case "b", "bet":
				cmdBet(irccon, command.Channel, command.Nick, command.Args)
			case "p", "poll":
				if !poll && !quiz {
					go cmdPoll(irccon, command.Channel, c, strings.Join(command.Args, " "))
				}
			case "pb", "processbets":
				cmdProcessBets(irccon, command.Channel, command.Nick)
			case "qz", "quiz":
				if !quiz && !poll {
					go cmdQuiz(irccon, command.Channel, c, strings.Join(command.Args, " "))
				}
			case "q", "quote":
				cmdQuote(irccon, command.Channel, command.Args)
			case "wbc", "points":
				go cmdStandings(irccon, command.Channel, command.Nick, "bet")
			case "wdc":
				go cmdStandings(irccon, command.Channel, command.Nick, "driver")
			case "wcc":
				go cmdStandings(irccon, command.Channel, command.Nick, "constructor")
			case "w", "weather":
				go cmdWeather(irccon, command.Channel, command.Nick, command.Args)
			}
		}
	})
	go tskEvents(irccon)
	go tskFeeds2(irccon)
	irccon.Loop()
}
