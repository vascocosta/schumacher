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
	"github.com/thoj/go-ircevent"
	"log"
	"math/rand"
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
		delta := time.Until(t)
		if delta >= 0 {
			event = []string{e[0], e[1], e[2], e[3], e[4]}
			return event, nil
		}
	}
	err = errors.New("No event found.")
	return
}

// The help command receives an IRC connection pointer, a channel and a search string.
// It then shows a compact help message listing all the possible commands of the bot.
func cmdHelp(irccon *irc.Connection, channel string, search string) {
	help := [9]string{
		prefix + "ask <question>.",
		prefix + "bet <xxx> <yyy> <zzz> - Place a bet for the next F1 race.",
		prefix + "help - Show this help message.",
		prefix + "next [category] - Show the next motorsport event.",
		prefix + "quiz [number] - Start an F1 quiz game.",
		prefix + "quote [get/add] [text] - Get a random quote or add one.",
		prefix + "wbc - Show the current Betting Championship standings.",
		prefix + "wcc - Show the current World Constructor Championship standings.",
		prefix + "wdc - Show the current World Driver Championship standings.",
	}
	for _, value := range help {
		if search == "" {
			irccon.Privmsg(channel, value)
		} else if strings.HasPrefix(value[1:], strings.ToLower(search)) {
			irccon.Privmsg(channel, value)
			return
		}
		time.Sleep(1 * time.Second)
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
	// Get the raw data through HTTP.
	data, err := getURL(url)
	if err != nil {
		irccon.Privmsg(channel, "Error getting standings.")
		log.Println("cmdStandings:", err)
		return
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
				output += fmt.Sprintf("%d. %s %d | ", i+1, score.Nick, score.Points)
			}
		}
		irccon.Privmsg(channel, output[:len(output)-3])
		return
	case "driver", "drivers":
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
		event, err = findNext("any", "any")
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
		irccon.Privmsg(channel, "You're not a registered user. Ask gluon for a bot account.")
		return
	}
	event, err := findNext("[formula 1]", "race")
	if err != nil {
		irccon.Privmsg(channel, "Bets are closed.")
		log.Println("cmdBet:", err)
		return
	}
	if len(bet) == 0 {
		bets, err = readCSV(betsFile)
		if err != nil {
			irccon.Privmsg(channel, "Error getting bets.")
			log.Println("cmdBet:", err)
			return
		}
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
	if len(bet) != 3 {
		irccon.Privmsg(channel, "The bet must contain 3 drivers.")
		return
	}
	drivers, err := readCSV(driversFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting drivers.")
		log.Println("cmdBet:", err)
		return
	}
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
	bets, err = readCSV(betsFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting bets.")
		log.Println("cmdBet:", err)
		return
	}
	for i := 0; i < len(bets); i++ {
		if strings.ToLower(bets[i][0]) == strings.ToLower(event[1]) && strings.ToLower(bets[i][1]) == strings.ToLower(nick) {
			update = true
			bets[i] = []string{strings.ToLower(event[1]), strings.ToLower(nick), first, second, third, "0"}
			break
		}
	}
	if !update {
		bets = append(bets, []string{strings.ToLower(event[1]), strings.ToLower(nick), first, second, third, "0"})
	}
	err = writeCSV(betsFile, bets)
	if err != nil {
		irccon.Privmsg(channel, "Error updating bet.")
		log.Println("cmdBet:", err)
		return
	}
	irccon.Privmsg(channel, "Your bet for the "+event[1]+" was successfully updated.")
}

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
	for i, bet := range bets {
		score := 0
		if strings.ToLower(bet[0]) == strings.ToLower(results[0][0]) {
			if strings.ToLower(bet[2]) == strings.ToLower(results[0][1]) {
				score += 10
			}
			if strings.ToLower(bet[3]) == strings.ToLower(results[0][2]) {
				score += 10
			}
			if strings.ToLower(bet[4]) == strings.ToLower(results[0][3]) {
				score += 10
			}
			bets[i][5] = strconv.Itoa(score)
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
	// This is an obfuscated way to randomise a slice that I googled in order to randomise the questions.
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(questions), func(i, j int) { questions[i], questions[j] = questions[j], questions[i] })
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
				i+1, n, questions[i][0], 20-time.Since(start).Seconds(),
			),
		)
		select {
		case answer := <-c:
			if strings.ToLower(answer[1]) == strings.ToLower(questions[i][1]) {
				timer.Reset(time.Duration(quizTimeout) * time.Second)
				start = time.Now()
				irccon.Privmsg(channel, "Correct!")
				score[answer[0]] += 1
			} else if answer[0] == nick && answer[1] == "--TIMEOUT--" {
				timer.Reset(time.Duration(quizTimeout) * time.Second)
				start = time.Now()
				irccon.Privmsg(channel, "Time's up... The correct answer was: "+questions[i][1])
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
	// If there are no arguments or if the first argument is "get", show a random quote.
	// We seed the randomizer with some variable number, the current time in nano seconds.
	// Then we set the index to the quotes to a random number between 0 and the length of quotes.
	// Finally we show a random quote on the channel.
	if len(args) == 0 || (len(args) > 0 && strings.ToLower(args[0]) == "get") {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(quotes))
		irccon.Privmsg(channel, fmt.Sprintf("%s - %s", quotes[index][1], quotes[index][0]))
		// If there is more than one argument and the first argument is "add", add the provided quote.
		// Finally we show a confirmation message on the channel.
	} else if len(args) > 1 && strings.ToLower(args[0]) == "add" {
		quotes = append(quotes, []string{time.Now().Format("02-01-2006"), strings.Join(args[1:], " ")})
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