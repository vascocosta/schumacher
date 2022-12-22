/*
 *  schumacher, a simple general purpose bot for IRC.
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
	"log"
	"strings"
	"time"

	irc "github.com/thoj/go-ircevent"
)

var nick = "Schumacher"                           // Nick to be used by the bot.
var channels = "#motorsport"                      // Names of the channels to join.
var adminNick = "gluon"                           // Nick used by the admin of the bot.
var inputFile = "/home/gluon/mnt/schumacher/in"   // File used to read messages.
var outputFile = "/home/gluon/mnt/schumacher/out" // File used to output messages.
var poll bool                                     // Bool to check if a poll is on.
var quiz bool                                     // Bool to check if a quiz is on.
var activeChannel string                          // The active channel.

// The main function handles flags, defines some IRC callbacks to handle events and launches background tasks.
func main() {
	flag.StringVar(&nick, "nick", "Schumacher", "Nick to be used by the bot.")
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
		// Every message, except the ones from the bot itself, are sent to an output file.
		// We try to parse a command from every PRIVMSG that the bot sees on each channel.
		// If we cannot parse a command, this means the message is just a regular message.
		// So we need to check if there's an ongoing poll or quiz or an embedded HTTP URL.
		// In case there's an ongoing poll or quiz, we send the nick/message to a channel.
		// Otherwise, if the message contains "http", we try to obtain its HTML title tag.
		// Finally, if we successfully parse a command, we call the matching cmd function.
		err = writeOut(outputFile, event.Arguments[0]+" "+event.Nick+" "+event.Message()+"\n")
		if err != nil {
			log.Println("main:", err)
		}
		m := strings.Trim(event.Message(), " ")
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
				/*
					case "rr", "register":
						cmdRegister(irccon, command.Channel, command.Nick)
					case "wbc", "points":
						go cmdStandings(irccon, command.Channel, command.Nick, "bet")
				*/
			case "wdc":
				go cmdStandings(irccon, command.Channel, command.Nick, "driver")
			case "wcc":
				go cmdStandings(irccon, command.Channel, command.Nick, "constructor")
			default:
				finishedCh := make(chan bool)
				go func() {
					select {
					case <-finishedCh:
					case <-time.After(4 * time.Second):
						irccon.Privmsg(command.Channel, "Command is taking long to run... Please wait.")
					}
				}()
				go cmdPlugin(strings.ToLower(command.Name), irccon, command.Channel, command.Nick, command.Args, finishedCh)
			}
		}
	})
	// Here we launch some background tasks that run in parallel with the main goroutine.
	go tskEvents(irccon)
	go tskFeeds(irccon)
	go tskWrite(irccon, inputFile)
	irccon.Loop()
}
