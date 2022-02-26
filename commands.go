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
	"encoding/json"
	"errors"
	"fmt"
	owm "github.com/briandowns/openweathermap"
	"github.com/thoj/go-ircevent"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The findNext function receives a category and session and returns the chronologically next event matching that criteria.
func findNext(category string, session string) (event []string, err error) {
	var t time.Time
	var timeFormat = "2006-01-02 15:04:05 UTC"
	events, err := readCSV(eventsFile)
	if err != nil {
		return
	}
	// Loop through all events and get a parsed time for the event that matches the category and session criteria.
	// There are 3 special cases where the category and session can be set to the wildcard any in different ways.
	// Otherwise, use the default case to search for a specific category and session.
	for _, e := range events {
		switch {
		case strings.ToLower(category) == "any" && strings.ToLower(session) == "any":
			t, err = time.Parse(timeFormat, e[3])
			if err != nil {
				err = errors.New("Error parsing time.")
				return event, err
			}
		case strings.ToLower(category) != "any" && strings.ToLower(session) == "any":
			if strings.ToLower(e[0]) == strings.ToLower(category) {
				t, err = time.Parse(timeFormat, e[3])
				if err != nil {
					err = errors.New("Error parsing time.")
					return event, err
				}
			}
		case strings.ToLower(category) == "any" && strings.ToLower(session) != "any":
			if strings.ToLower(e[2]) == strings.ToLower(session) {
				t, err = time.Parse(timeFormat, e[3])
				if err != nil {
					err = errors.New("Error parsing time.")
					return event, err
				}
			}
		default:
			if strings.ToLower(e[0]) == strings.ToLower(category) && strings.ToLower(e[2]) == strings.ToLower(session) {
				t, err = time.Parse(timeFormat, e[3])
				if err != nil {
					err = errors.New("Error parsing time.")
					return event, err
				}
			}
		}
		// Get the time delta from now until the time of the event.
		// If delta is equal or greater than zero, this is the next event that will happen.
		delta := time.Until(t)
		if delta >= 0 {
			event = []string{e[0], e[1], e[2], e[3], e[4], e[5]}
			return event, nil
		}
	}
	err = errors.New("No event found.")
	return
}

// The help command receives an IRC connection pointer, a channel and a search string.
// It then shows a compact help message listing all the possible commands of the bot.
func cmdHelp(irccon *irc.Connection, channel string, search string) {
	help := [13]string{
		"ask <question>.",
		"bet <xxx> <yyy> <zzz> or [log/odds/nick] - Place a bet for the next F1 race or get bet info.",
		"help - Show this help message.",
		"next [category] - Show the next motorsport event.",
		"notify [on/off] - Turn on/off notifications for the current channel.",
		"omdb [movie/show] - Show info about a movie or a show.",
		"poll <question;option_1;option_2;option_n>",
		"quiz [number] - Start an F1 quiz game.",
		"quote [get/add] [text] - Get a random quote or add one.",
		"wbc - Show the current Betting Championship standings.",
		"wcc - Show the current World Constructor Championship standings.",
		"wdc - Show the current World Driver Championship standings.",
		"weather [location] - Show the current weather for a locattion.",
	}
	if search == "" {
		var commandList string
		irccon.Privmsg(channel, "This is a list of all the commands of this bot, !help command_name shows how to use each one:")
		for _, v := range help {
			commandList += prefix+strings.Split(v, " ")[0]+" "
		}
		irccon.Privmsg(channel, commandList)
	} else {
		for _, v := range help {
			if strings.HasPrefix(v, strings.ToLower(search)) {
				irccon.Privmsg(channel, prefix+v)
				return
			}
		}
	}
}

// The standings command receives an IRC connection pointer, a channel, a nick and a championship string.
// It then queries the Ergast F1 API for either the WDC or the WCC and displays the results on the channel.
func cmdStandings(irccon *irc.Connection, channel string, nick string, championship string) {
	var output string
	// Base URL of the Ergast F1 API.
	url := "http://ergast.com/api/f1/current/"
	if strings.ToLower(championship) == "constructor" || strings.ToLower(championship) == "constructors" {
		championship = "constructor"
		url += "constructorStandings.json"
	} else if strings.ToLower(championship) == "driver" || strings.ToLower(championship) == "drivers" {
		championship = "driver"
		url += "driverStandings.json"
	} else {
		championship = "bet"
	}
	// Display the results on the channel, depending on which kind of championship was requested.
	// The API returns in JSON format which is decoded to either a DStandings or CStandings strut.
	switch strings.ToLower(championship) {
	case "bet", "bets":
		users, err := readCSV(usersFile)
		if err != nil {
			irccon.Privmsg(channel, "Error getting users.")
			log.Println("cmdStandings:", err)
			return
		}
		scoreList := make(ScoreList, len(users))
		for _, user := range users {
			points, _ := strconv.Atoi(user[2])
			if points > 0 {
				re, err := regexp.Compile("[^a-zA-Z0-9]+")
				if err != nil {
					irccon.Privmsg(channel, "Error getting standings.")
					log.Println("cmdStandings:", err)
					return
				}
				scoreList = append(scoreList, Score{strings.ToUpper(re.ReplaceAllString(user[0], "")[0:3]), points})
			}
		}
		sort.Sort(sort.Reverse(scoreList))
		for i, score := range scoreList {
			if score.Points > 0 {
				output += fmt.Sprintf("%d. [%s]: %d | ", i+1, score.Nick, score.Points)
			}
		}
		if len(output) > 3 {
			irccon.Privmsg(channel, output[:len(output)-3])
		}
		return
	case "driver", "drivers":
		// Get the raw data through HTTP.
		data, err := getURL(url)
		if err != nil {
			irccon.Privmsg(channel, "Error getting standings.")
			log.Println("cmdStandings:", err)
			return
		}
		var standings DStandings
		err = json.Unmarshal(data, &standings)
		if err != nil {
			irccon.Privmsg(channel, "Error getting driver standings.")
			log.Println("cmdStandings:", err)
			return
		}
		for _, driver := range standings.MRData.StandingsTable.StandingsLists[0].DriverStandings {
			output += fmt.Sprintf(
				"%s. %s %s (%s wins) ",
				driver.Position,
				driver.Driver.Code,
				driver.Points,
				driver.Wins,
			)
		}
	case "constructor", "constructors":
		// Get the raw data through HTTP.
		data, err := getURL(url)
		if err != nil {
			irccon.Privmsg(channel, "Error getting standings.")
			log.Println("cmdStandings:", err)
			return
		}
		var standings CStandings
		err = json.Unmarshal(data, &standings)
		if err != nil {
			irccon.Privmsg(channel, "Error getting constructor standings.")
			log.Println("cmdStandings:", err)
			return
		}
		for _, constructor := range standings.MRData.StandingsTable.StandingsLists[0].ConstructorStandings {
			output += fmt.Sprintf(
				"%s. %s %s (%s wins) ",
				constructor.Position,
				constructor.Constructor.Name,
				constructor.Points,
				constructor.Wins,
			)
		}
	}
	irccon.Privmsg(channel, output)
}

// The next command receives an IRC connection pointer, a channel, a nick and an optional search string.
// It then queries the events CSV file and returns which event is happening next, showing it on the channel.
func cmdNext(irccon *irc.Connection, channel string, nick string, search string) {
	var tz = "Europe/Berlin"
	var event []string
	var timeFormat = "2006-01-02 15:04:05 UTC"
	users, err := readCSV(usersFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting users.")
		log.Println("cmdNext:", err)
		return
	}
	for _, user := range users {
		if strings.ToLower(user[0]) == strings.ToLower(nick) {
			tz = user[1]
		}
	}
	// Do some search string replacements in case there's actually a search argument.
	// Users use abreviated search terms, which are expanded for better database matching.
	// Retrieve the next event matching category or session criteria.
	// Else, simply retrieve the next event from any category or session type.
	if search != "" {
		switch search {
		case "f1", "formula1":
			event, err = findNext("[Formula 1]", "any")
		case "f2", "formula2":
			event, err = findNext("[Formula 2]", "any")
		case "f3", "formula3":
			event, err = findNext("[Formula 3]", "any")
		case "q", "quali", "qualy", "qualifier", "qualifying":
			event, err = findNext("[Formula 1]", "Qualifying")
		case "r", "race":
			event, err = findNext("[Formula 1]", "Race")
		case "s", "sprint":
			event, err = findNext("[Formula 1]", "Sprint Race")
		default:
			event, err = findNext("["+search+"]", "any")
		}
	} else {
		switch channel {
		case "#formula1":
			event, err = findNext("[Formula 1]", "any")
		case "#geeks":
			event, err = findNext("[Space]", "any")
		default:
			event, err = findNext("any", "any")
		}
	}
	if err != nil {
		irccon.Privmsg(channel, "No event found.")
		log.Println("cmdNext:", err)
		return
	}
	// Parse the time of the event, calculate time delta, do some formatting and finally show the results.
	// The times are localised as per the user's time zone before being shown.
	// The time delta between now and the next event uses modulo to perfectly round days, hour an minutes.
	t, err := time.Parse(timeFormat, event[3])
	if err != nil {
		irccon.Privmsg(channel, "Error parsing time.")
		log.Println("cmdNext: Error parsing time.")
		return
	}
	delta := time.Until(t)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		irccon.Privmsg(channel, "Error converting time to user time zone. Using default one.")
		log.Println("cmdNext: Error converting time to user time zone. Using default one.")
		loc, _ = time.LoadLocation("Europe/Berlin")
	}
	t = t.In(loc)
	wday := t.Weekday().String()
	mday := t.Day()
	month := t.Month()
	hour := t.Hour()
	min := t.Minute()
	zone, offset := t.Zone()
	uoffset := offset / 3600
	delta = delta / 1000000000
	days := int((delta % (86400 * 30)) / 86400)
	hours := int((delta % 86400) / 3600)
	minutes := int((delta % 3600) / 60)
	irccon.Privmsg(channel, fmt.Sprintf(
		"%s, %d %s at %02d:%02d \x02%s (UTC+%d)\x02 | %s | %d day(s), %d hour(s), %d minute(s)",
		wday, mday, month, hour, min, zone, uoffset, event[0]+" "+event[1]+" "+event[2], days, hours, minutes))
}

// The bet command receives an IRC connection pointer, a channel, a nick and a bet containing 3 drivers.
// It then stores the bet provided by the user, or lets the user know his current bet for the next race.
func cmdBet(irccon *irc.Connection, channel string, nick string, bet []string) {
	var correct int
	var bets [][]string
	var update bool
	users, err := readCSV(usersFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting users.")
		log.Println("cmdBet:", err)
		return
	}
	if !isUser(nick, users) {
		irccon.Privmsg(channel, "You're not a registered user. Use !register to register your nick.")
		return
	}
	event, err := findNext("[formula 1]", "race")
	if err != nil {
		irccon.Privmsg(channel, "Bets are closed.")
		log.Println("cmdBet:", err)
		return
	}
	bets, err = readCSV(betsFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting bets.")
		log.Println("cmdBet:", err)
		return
	}
	// If no bet is provided as argument, we simply show the user's current bet, if he's placed one.
	if len(bet) == 0 {
		for i := len(bets) - 1; i >= 0; i-- {
			if strings.ToLower(bets[i][0]) == strings.ToLower(event[1]) && strings.ToLower(bets[i][1]) == strings.ToLower(nick) {
				first := strings.ToUpper(bets[i][2])
				second := strings.ToUpper(bets[i][3])
				third := strings.ToUpper(bets[i][4])
				irccon.Privmsg(channel, fmt.Sprintf("Your current bet for the %s: %s %s %s", event[1], first, second, third))
				return
			}
		}
		irccon.Privmsg(channel, fmt.Sprintf("You haven't placed a bet for the %s yet.", event[1]))
		return
	}
	drivers, err := readCSV(driversFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting drivers.")
		log.Println("cmdBet:", err)
		return
	}
	// If instead of a normal bet the user provides a single word, we interpret it as an argument.
	// There are multiple arguments, for which we show the driver odds or a log of the user's bets.
	// Other possible argument is the nick of a registered user, for which we show that user's bet.
	// Alternatively, if the argument provided isn't a valid command option, we let the user know.
	if len(bet) == 1 {
		switch strings.ToLower(bet[0]) {
		case "multipliers", "odds":
			var output string
			odds, err := toStringMap(drivers, 1, 2)
			if err != nil {
				irccon.Privmsg(channel, "Error getting odds.")
				log.Println("cmdBet:", err)
				return
			}
			for k, v := range odds {
				output += fmt.Sprintf("[%s]: %s | ", strings.ToUpper(k), v)
			}
			if len(output) > 3 {
				irccon.Privmsg(channel, output[:len(output)-3])
			}
		case "log":
			var betsFound bool
			var counter int
			for i := len(bets) - 1; i >= 0 && counter < 3; i-- {
				if strings.ToLower(bets[i][1]) == strings.ToLower(nick) {
					betsFound = true
					irccon.Privmsg(channel,
						fmt.Sprintf("Your bet for the %s: %s %s %s %s points.",
							bets[i][0],
							strings.ToUpper(bets[i][2]),
							strings.ToUpper(bets[i][3]),
							strings.ToUpper(bets[i][4]),
							bets[i][5]))
					counter += 1
				}
			}
			if !betsFound {
				irccon.Privmsg(channel, "No recent bets from you.")
			}
		default:
			if isUser(strings.ToLower(bet[0]), users) {
				for i := len(bets) - 1; i >= 0; i-- {
					if strings.ToLower(bets[i][0]) == strings.ToLower(event[1]) &&
						strings.ToLower(bets[i][1]) == strings.ToLower(bet[0]) {
						first := strings.ToUpper(bets[i][2])
						second := strings.ToUpper(bets[i][3])
						third := strings.ToUpper(bets[i][4])
						irccon.Privmsg(channel,
							fmt.Sprintf("%s's current bet for the %s: %s %s %s",
								bet[0],
								event[1],
								first,
								second,
								third))
						return
					}
				}
				irccon.Privmsg(channel, "That user hasn't bet for the current race yet.")
				return
			}
			irccon.Privmsg(channel, "Unknown command option.")
		}
		return
	}
	if len(bet) != 3 {
		irccon.Privmsg(channel, "The bet must contain 3 drivers.")
		return
	}
	// Finally, if we reach this point, it means the user has provided a valid bet composed of 3 drivers.
	// We verify that all 3 driver codes are valid as per the drivers CSV file before we go any further.
	// If the 3 codes are valid, we either place a new bet or update an already placed bet for the race.
	first := strings.ToLower(bet[0])
	second := strings.ToLower(bet[1])
	third := strings.ToLower(bet[2])
	for _, driver := range drivers {
		code := strings.ToLower(driver[1])
		if code == first || code == second || code == third {
			correct++
		}
	}
	if correct != 3 {
		irccon.Privmsg(channel, "Invalid drivers.")
		return
	}
	for i := 0; i < len(bets); i++ {
		if strings.ToLower(bets[i][0]) == strings.ToLower(event[1]) && strings.ToLower(bets[i][1]) == strings.ToLower(nick) {
			update = true
			bets[i] = []string{event[1], strings.ToLower(nick), first, second, third, "0"}
			break
		}
	}
	if !update {
		bets = append(bets, []string{event[1], strings.ToLower(nick), first, second, third, "0"})
	}
	err = writeCSV(betsFile, bets)
	if err != nil {
		irccon.Privmsg(channel, "Error updating bet.")
		log.Println("cmdBet:", err)
		return
	}
	irccon.Privmsg(channel, "Your bet for the "+event[1]+" was successfully updated.")
}

// The processbets command receives an IRC connection pointer, a channel and a nick.
// It then processes the placed bets, according to the results in the results file.
func cmdProcessBets(irccon *irc.Connection, channel string, nick string) {
	if strings.ToLower(nick) != strings.ToLower(adminNick) {
		irccon.Privmsg(channel, "Only "+adminNick+" can use this command.")
		return
	}
	results, err := readCSV(resultsFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting results.")
		log.Println("cmdProcessBets:", err)
		return
	}
	if results[0][0] == results[0][4] {
		irccon.Privmsg(channel, results[0][0]+" bets have already been processed in the past.")
		return
	}
	bets, err := readCSV(betsFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting bets.")
		log.Println("cmdProcessBets:", err)
		return
	}
	users, err := readCSV(usersFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting users.")
		log.Println("cmdProcessBets:", err)
		return
	}
	drivers, err := readCSV(driversFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting drivers.")
		log.Println("cmdProcessBets:", err)
		return
	}
	odds, err := toStringMap(drivers, 1, 2)
	if err != nil {
		irccon.Privmsg(channel, "Error getting odds.")
		log.Println("cmdProcessBets:", err)
		return
	}
	// This is the main loop where we go through each bet placed by the user and process it.
	// If the race on the bet matches the race on the results file, we calculate its score.
	for i, bet := range bets {
		score := 0
		first := strings.ToLower(results[0][1])
		second := strings.ToLower(results[0][2])
		third := strings.ToLower(results[0][3])
		if strings.ToLower(bet[0]) == strings.ToLower(results[0][0]) {
			// If the first driver is on the podium, we have two different possibilities.
			// If the first driver is the first on the results, we score 10 * multiplier.
			// If the first driver is not the first on the results, we score 5 * multiplier.
			if contains([]string{first, second, third}, strings.ToLower(bet[2])) {
				multiplier, err := strconv.Atoi(odds[bet[2]])
				if err != nil {
					irccon.Privmsg(channel, "Error applying multiplier.")
					log.Println("cmdProcessBets:", err)
					return
				}
				if strings.ToLower(bet[2]) == strings.ToLower(results[0][1]) {
					score += (10 * multiplier)
				} else {
					score += (5 * multiplier)
				}
			}
			// If the second driver is on the podium, we have two different possibilities.
			// If the second driver is the second on the results, we score 10 * multiplier.
			// If the second driver is not the second on the results, we score 5 * multiplier.
			if contains([]string{first, second, third}, strings.ToLower(bet[3])) {
				multiplier, err := strconv.Atoi(odds[bet[3]])
				if err != nil {
					irccon.Privmsg(channel, "Error applying multiplier.")
					log.Println("cmdProcessBets:", err)
					return
				}
				if strings.ToLower(bet[3]) == strings.ToLower(results[0][2]) {
					score += (10 * multiplier)
				} else {
					score += (5 * multiplier)
				}
			}
			// If the third driver is on the podium, we have two different possibilities.
			// If the third driver is the third on the results, we score 10 * multiplier.
			// If the third driver is not the third on the results, we score 5 * multiplier.
			if contains([]string{first, second, third}, strings.ToLower(bet[4])) {
				multiplier, err := strconv.Atoi(odds[bet[4]])
				if err != nil {
					irccon.Privmsg(channel, "Error applying multiplier.")
					log.Println("cmdProcessBets:", err)
					return
				}
				if strings.ToLower(bet[4]) == strings.ToLower(results[0][3]) {
					score += (10 * multiplier)
				} else {
					score += (5 * multiplier)
				}
			}
			bets[i][5] = strconv.Itoa(score)
			// Update the total number of points for each driver on the users file.
			// The code above only handles points for each bet, not for each user.
			for j, user := range users {
				if strings.ToLower(user[0]) == strings.ToLower(bet[1]) {
					currentScore, err := strconv.Atoi(users[j][2])
					if err != nil {
						irccon.Privmsg(channel, "Error getting current score.")
						log.Println("cmdProcessBets:", err)
						return
					}
					users[j][2] = strconv.Itoa(currentScore + score)
				}
			}
		}
		err = writeCSV(usersFile, users)
		if err != nil {
			irccon.Privmsg(channel, "Error storing user points.")
			log.Println("cmdProcessBets:", err)
			return
		}
	}
	// Finally update the bets file with the points for each bet for the current race.
	// The results file is updated so that the last field is set to the current race.
	err = writeCSV(betsFile, bets)
	if err != nil {
		irccon.Privmsg(channel, "Error storing bet points.")
		log.Println("cmdProcessBets:", err)
		return
	}
	results[0][4] = results[0][0]
	err = writeCSV(resultsFile, results)
	if err != nil {
		irccon.Privmsg(channel, "Error storing last processed bet..")
		log.Println("cmdProcessBets:", err)
		return
	}
	irccon.Privmsg(channel, results[0][0]+" bets successfully processed.")
}

// The poll command receives an IRC connection pointer, an IRC channel, a [2]string channel and poll data.
// It runs as a goroutine that makes a poll on an IRC channel using the given poll data.
// It then waits for votes from the users and finally displays the results of the poll.
func cmdPoll(irccon *irc.Connection, channel string, c chan [2]string, pollData string) {
	// Parse the poll data into a question and possible answers and then show it on the IRC channel.
	parsed := strings.Split(pollData, ";")
	if len(parsed) <= 1 {
		irccon.Privmsg(channel, "Syntax: !poll question;option 1;option 2;option n")
		return
	}
	irccon.Privmsg(channel, fmt.Sprintf("Poll: %s (%d seconds to vote)", parsed[0], pollTimeout))
	time.Sleep(1 * time.Second)
	for k, v := range parsed[1:] {
		irccon.Privmsg(channel, fmt.Sprintf("%d. %s", k+1, v))
		time.Sleep(1 * time.Second)
	}
	poll = true
	activeChannel = channel
	votes := make(map[string]int)
	results := make(map[string]int)
	var total int
	// This is the main loop of the goroutine, which waits for answers to the poll on the "c" go channel.
	// The answer is sent to the "c" go channel on the main goroutine, inside the "PRIVMSG" callback.
	// Eventually it times out after pollTimeout seconds and shows the results of the poll on the IRC channel.
	time.AfterFunc(time.Duration(pollTimeout)*time.Second, func() {
		c <- [2]string{nick, "--TIMEOUT--"}
	})
	for answer := range c {
		if answer[0] == nick && answer[1] == "--TIMEOUT--" {
			poll = false
			activeChannel = ""
			irccon.Privmsg(channel, "The Poll has ended.")
			if len(votes) > 0 {
				time.Sleep(1 * time.Second)
				irccon.Privmsg(channel, "Results: ")
				for _, v := range votes {
					results[strconv.Itoa(v)] += 1
				}
				for _, v := range results {
					total += v
				}
				for k, v := range results {
					index, _ := strconv.Atoi(k)
					irccon.Privmsg(channel,
						fmt.Sprintf("%s. %s - %.2f%% votes",
							k,
							parsed[index],
							(float32(v)/float32(total))*100))
				}
			}
			return
		} else {
			vote, err := strconv.Atoi(answer[1])
			if err == nil && (vote > 0 && vote <= len(parsed[1:])) {
				votes[answer[0]] = vote
			}
		}
	}
}

// The quiz command receives an IRC connection pointer, an IRC channel and a [2]string channel.
// It runs as a goroutine that opens a quiz file and asks questions on the given IRC channel.
// It then waits for answers to classify as correct or wrong or times out after a while.
func cmdQuiz(irccon *irc.Connection, channel string, c chan [2]string, number string) {
	quiz = true
	activeChannel = channel
	score := make(map[string]int)
	n, err := strconv.Atoi(number)
	if err != nil || (n <= 0 || n > 10) {
		n = 5
	}
	questions, err := readCSV(quizFile)
	if err != nil {
		irccon.Privmsg(channel, "Error reading questions.")
		log.Println("cmdQuiz:", err)
		activeChannel = ""
		return
	}
	// Filter only the questions of the current channel.
	var channelQuestions [][]string
	for _, question := range questions {
		if strings.ToLower(question[2]) == strings.ToLower(channel) {
			channelQuestions = append(channelQuestions, question)
		}
	}
	// This is an obfuscated way to randomise a slice that I googled in order to randomise the questions.
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(channelQuestions), func(i, j int) { channelQuestions[i], channelQuestions[j] = channelQuestions[j], channelQuestions[i] })
	// This is the main loop of the goroutine, which for now only asks up to 5 questions to avoid SPAM.
	// After showing the question on irc, it waits for an answer on the "c" go channel and classifies it.
	// The answer is sent to the "c" go channel on the main goroutine, inside the "PRIVMSG" callback.
	// Eventually if no correct answer is sent, it times out after quizTimeout seconds.
	timer := time.AfterFunc(time.Duration(quizTimeout)*time.Second, func() {
		c <- [2]string{nick, "--TIMEOUT--"}
	})
	start := time.Now()
	for i := 0; i < n; i++ {
		irccon.Privmsg(
			channel,
			fmt.Sprintf(
				"%d/%d - %s (%0.0f seconds remaining)",
				i+1, n, channelQuestions[i][0], 20-time.Since(start).Seconds(),
			),
		)
		select {
		case answer := <-c:
			if strings.ToLower(answer[1]) == strings.ToLower(channelQuestions[i][1]) {
				timer.Reset(time.Duration(quizTimeout) * time.Second)
				start = time.Now()
				irccon.Privmsg(channel, "Correct!")
				score[answer[0]] += 1
			} else if answer[0] == nick && answer[1] == "--TIMEOUT--" {
				timer.Reset(time.Duration(quizTimeout) * time.Second)
				start = time.Now()
				irccon.Privmsg(channel, "Time's up... The correct answer was: "+channelQuestions[i][1])
			} else {
				if i >= 0 {
					i-- // Avoid advancing to the next question, when answer is wrong.
				}
				irccon.Privmsg(channel, "Wrong!")
			}
		}
	}
	// At the end of the quiz we stop the timer and set quiz to false.
	// We then proceed to show the final score sorted by points (value).
	// Internally we store the score in a map[string]int but fmt only sorts maps by key.
	// Therefore we must use a trick to sort by points (value) which is to use sort.Sort.
	// sort.Sort requires us to use a slice of struts (ScoreList) which we declare above.
	// We create a ScoreList with the length of scores and populate it with its values.
	// Finally we use sort.Reverse to sort by highest score and show the results.
	timer.Stop()
	quiz = false
	activeChannel = ""
	irccon.Privmsg(channel, "The quiz is over!")
	time.Sleep(1 * time.Second)
	irccon.Privmsg(channel, "Score:")
	scoreList := make(ScoreList, len(score))
	i := 0
	for key, value := range score {
		scoreList[i] = Score{key, value}
		i++
	}
	sort.Sort(sort.Reverse(scoreList))
	for _, value := range scoreList {
		irccon.Privmsg(channel, fmt.Sprintf("%s - %d", value.Nick, value.Points))
		time.Sleep(1 * time.Second)
	}
}

// The quote command receives an IRC connection pointer, a channel and an arguments slice of strings.
// It then checks if there are arguments and displays a random quote or adds a new quote accordingly.
func cmdQuote(irccon *irc.Connection, channel string, args []string) {
	// Get a collection of quotes stored as a CSV file.
	quotes, err := readCSV(quotesFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting quote.")
		log.Println("cmdQuote:", err)
		return
	}
	// Filter only the quotes of the current channel.
	var channelQuotes [][]string
	for _, quote := range quotes {
		if strings.ToLower(quote[2]) == strings.ToLower(channel) {
			channelQuotes = append(channelQuotes, quote)
		}
	}
	// If there are no arguments or if the first argument is "get", show a random quote.
	// We seed the randomizer with some variable number, the current time in nano seconds.
	// Then we set the index to the quotes to a random number between 0 and the length of quotes.
	// Finally we show a random quote on the channel.
	if len(args) == 0 || (len(args) > 0 && strings.ToLower(args[0]) == "get") {
		if len(channelQuotes) == 0 {
			irccon.Privmsg(channel, "There are no quotes for this channel.")
			return
		}
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(channelQuotes))
		irccon.Privmsg(channel, fmt.Sprintf("%s - %s", channelQuotes[index][1], channelQuotes[index][0]))
		// If there is more than one argument and the first argument is "add", add the provided quote.
		// Finally we show a confirmation message on the channel.
	} else if len(args) > 1 && strings.ToLower(args[0]) == "add" {
		quotes = append(quotes, []string{time.Now().Format("02-01-2006"), strings.Join(args[1:], " "), strings.ToLower(channel)})
		err = writeCSV(quotesFile, quotes)
		if err != nil {
			irccon.Privmsg(channel, "Error adding quote.")
			log.Println("cmdQuote:", err)
			return
		}
		irccon.Privmsg(channel, "Quote added.")
		// Otherwise, if we get here, it means the user didn't use the command correctly.
		// Ttherefore we show a usage message on the channel.
	} else {
		irccon.Privmsg(channel, "Usage: !quote [get|add] [text]")
	}
}

// The ask command receives an IRC connection pointer, a channel and an arguments slice of strings.
// It then checks if the user has asked a question and displays a random answer on the channel.
func cmdAsk(irccon *irc.Connection, channel string, args []string) {
	// Get a collection of answers stored as a CSV file.
	answers, err := readCSV(answersFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting answer.")
		log.Println("cmdAsk:", err)
		return
	}
	// If the number of arguments is greater than 0, a question was asked, we show a random answer.
	// We seed the randomizer with some variable number, the current time in nano seconds.
	// Then we set the index to the answers to a random number between 0 and the length of answers.
	// Finally we show a random answer on the channel.
	if len(args) > 0 {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(answers))
		irccon.Privmsg(channel, fmt.Sprintf("%s", answers[index][0]))
		// Otherwise, if we get here, it means the user didn't use the command correctly.
		// Ttherefore we show a usage message on the channel.
	} else {
		irccon.Privmsg(channel, "Usage: !ask <question>")
	}
}

// The notify command receives an IRC connection pointer, a channel, a nick and an arguments slice of strings.
// It then enables or disables notifications for the current channel's events if the argument is on or off.
func cmdNotify(irccon *irc.Connection, channel string, nick string, args []string) {
	users, err := readCSV(usersFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting users.")
		log.Println("cmdNotify:", err)
		return
	}
	// If the number of arguments is exactly one, we check whether it is on or off.
	// If it is on, we add the current channel to the user's notification list.
	// If it is off, we remove the current channel from the user's notification list.
	if len(args) == 1 {
		if strings.ToLower(args[0]) == "on" {
			for i, user := range users {
				if strings.ToLower(user[0]) == strings.ToLower(nick) {
					channels := strings.Split(user[3], ":")
					if !contains(channels, channel) {
						channels = append(channels, channel)
						users[i][3] = strings.Trim(strings.Join(channels, ":"), ":")
						err = writeCSV(usersFile, users)
						if err != nil {
							irccon.Privmsg(channel, "Error storing notifications.")
							log.Println("cmdNotify:", err)
							return
						}
					}
					irccon.Privmsg(channel, "Notifications updated. Get mentions for events on: "+users[i][3])
				}
			}
		} else if strings.ToLower(args[0]) == "off" {
			for i, user := range users {
				if strings.ToLower(user[0]) == strings.ToLower(nick) {
					channels := strings.Split(user[3], ":")
					var updatedChannels string
					if contains(channels, channel) {
						for _, v := range channels {
							if v != channel {
								updatedChannels += v + ":"
							}
						}
						users[i][3] = strings.Trim(updatedChannels, ":")
						err = writeCSV(usersFile, users)
						if err != nil {
							irccon.Privmsg(channel, "Error storing notifications.")
							log.Println("cmdNotify:", err)
							return
						}
					}
					irccon.Privmsg(channel, "Notifications updated. Get mentions for events on: "+users[i][3])
				}
			}
		} else {
			irccon.Privmsg(channel, "Usage: !notifiy <on/off>")
		}
	} else {
		irccon.Privmsg(channel, "Usage: !notify <on/off>")
	}
}

// The weather command receives an IRC connection pointer, a channel, a nick and an arguments slice of strings.
// It then shows the current weather for a given location on the channel using the OpenWeatherMap API.
func cmdWeather(irccon *irc.Connection, channel string, nick string, args []string) {
	weather, err := readCSV(weatherFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting weather settings.")
		log.Println("cmdWeather:", err)
		return
	}
	location := ""
	tempUnits := "C"
	windUnits := "m/s"
	// Neither a location nor temperature unit were provided as an argument to the command.
	// So we must get the location and temperature unit for the user from the weather file.
	// If a user in the weather file matches nick, we get its location and temperature unit.
	if len(args) == 0 {
		for _, v := range weather {
			if strings.ToLower(v[0]) == strings.ToLower(nick) {
				tempUnits = strings.ToUpper(v[1])
				location = v[2]
			}
		}
		// A temperature unit was provided as an argument to the command, we must update the setting.
		// However, we must first check if the user already has a location set on the weather file.
		// If so, we update the user units, otherwise we ask him to get the wether for a location.
		// This is so that the user gets registered on the weather file before we can set a location.
	} else if len(args) == 1 && (strings.ToLower(args[0]) == "c" || strings.ToLower(args[0]) == "f") {
		var unitsUpdated bool
		for i, v := range weather {
			// User with a location on the weather database.
			if strings.ToLower(v[0]) == strings.ToLower(nick) {
				unitsUpdated = true
				weather[i][1] = strings.ToLower(args[0])
			}
		}
		if !unitsUpdated {
			irccon.Privmsg(channel, "Get the weather for some location before setting the units.")
			return
		}
		err = writeCSV(weatherFile, weather)
		if err != nil {
			irccon.Privmsg(channel, "Error storing weather units.")
			log.Println("cmdWeather:", err)
			return
		}
		irccon.Privmsg(channel, "Temperature units updated.")
		return
		// If we reach this point, a location was provided as an argument to the command.
		// If the user already exists, we update his location, otherwise we register him.
	} else {
		var newUser bool = true
		location = strings.Join(args, " ")
		for i, v := range weather {
			// User with a location on the weather database.
			if strings.ToLower(v[0]) == strings.ToLower(nick) {
				newUser = false
				weather[i][2] = location
			}
		}
		if newUser {
			// User without a location on the weather database.
			weather = append(weather, []string{nick, "c", location})
		}
		err = writeCSV(weatherFile, weather)
		if err != nil {
			irccon.Privmsg(channel, "Error storing weather location.")
			log.Println("cmdWeather:", err)
			return
		}
	}
	if location == "" {
		irccon.Privmsg(channel, "Please provide a location as argument.")
		return
	}
	if tempUnits == "F" {
		windUnits = "mph"
	}
	// Finally we get the current weather at a location using the temperature units.
	// Then we display a nicely formatted and compact weather string on the channel.
	w, err := owm.NewCurrent(tempUnits, "en", owmAPIKey)
	if err != nil {
		irccon.Privmsg(channel, "Error fetching weather.")
		log.Println("cmdWeather:", err)
		return
	}
	err = w.CurrentByName(location)
	if err != nil {
		irccon.Privmsg(channel, "Could not fetch weather for that location.")
		log.Println("cmdWeather:", err)
		return
	}
	irccon.Privmsg(
		channel,
		fmt.Sprintf("%s: %s | Temperature: %0.1f%s | Humidity: %d%% | Pressure: %0.1fhPa | Wind: %0.1f%s",
			w.Name,
			w.Weather[0].Description,
			w.Main.Temp,
			tempUnits,
			w.Main.Humidity,
			w.Main.Pressure,
			w.Wind.Speed,
			windUnits))
}

// The register command receives an IRC connection pointer, a channel and a nick.
// It then checks if the user isn't already registered and registers it with the bot.
func cmdRegister(irccon *irc.Connection, channel string, nick string) {
	users, err := readCSV(usersFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting users.")
		log.Println("cmdRegister:", err)
		return
	}
	// If the nick is already a known user to the bot, we don't register it.
	// Otherwise we add this new nick as a registered user on the users file.
	if isUser(strings.ToLower(nick), users) {
		irccon.Privmsg(channel, "Your nick is already registered.")
		return
	}
	users = append(users, []string{strings.ToLower(nick), "Europe/Berlin", "0", ""})
	err = writeCSV(usersFile, users)
	if err != nil {
		irccon.Privmsg(channel, "Error registering user.")
		log.Println("cmdRegister:", err)
		return
	}
	irccon.Privmsg(channel, "Your nick was successfully registered.")
}

// The plugin command receives a name, an IRC connection pointer, a channel, a nick and an arguments slice of strings.
// It then tries to execute the given plugin name if a file with that name is found on the plugins folder.
func cmdPlugin(name string, irccon *irc.Connection, channel string, nick string, args []string) {
	var cmd *exec.Cmd
	if !fileExists(pluginsFolder + name) {
		irccon.Privmsg(channel, "Unknown command or plugin.")
		return
	}
	if len(args) == 0 {
		cmd = exec.Command(pluginsFolder+name, nick)
	} else {
		cmd = exec.Command(pluginsFolder+name, nick, strings.Join(args, " "))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		irccon.Privmsg(channel, "Error executing plugin.")
		log.Println("cmdPlugin:", err)
		return
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		irccon.Privmsg(channel, line)
		time.Sleep(1 * time.Second)
	}
}
