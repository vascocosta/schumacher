/*
 *  schumacher, the irc bot of the #formula1 channel at Quakenet.
 *  Copyright (C) 2021  Vasco Costa
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
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	server       = "irc.quakenet.org:6667"                           // Hostname of the server to connect to.
	prefix       = "!"                                               // Prefix which is used by the user to issue commands.
	betsFile     = "/home/gluon/var/irc/bots/Schumacher/bets.csv"    // Full path to the bets file.
	driversFile  = "/home/gluon/var/irc/bots/Schumacher/drivers.csv" // Full path to the drivers file.
	eventsFile   = "/home/gluon/var/irc/bots/Schumacher/events.csv"  // Full path to the events file.
	feedsFile    = "/home/gluon/var/irc/bots/Schumacher/feeds.csv"   // Full path to the feeds file.
	usersFile    = "/home/gluon/var/irc/bots/Schumacher/users.csv"   // Full path to the users file.
	quizFile     = "/home/gluon/var/irc/bots/Schumacher/quiz.csv"    // Full path to the quiz file.
	quotesFile   = "/home/gluon/var/irc/bots/Schumacher/quotes.csv"  // Full path to the quotes file.
	quizTimeout  = 20
	hns          = 3600000000000
	feedInterval = 300
)

var nick = "Schumacher_"     // Nick to be used by the bot.
var channels = "#motorsport" // Names of the channels to join.
var quiz bool

// Type that represents an IRC command issued by the user.
type Command struct {
	Name    string
	Args    []string
	Nick    string
	Channel string
}

// Type that represents the F1 World Driver Championship standings.
type DStandings struct {
	MRData struct {
		XMLNS          string `json:"xmlns"`
		Series         string `json:"series"`
		URL            string `json:"url"`
		Limit          string `json:"limit"`
		Offset         string `json:"offset"`
		Total          string `json:"total"`
		StandingsTable struct {
			Season         string `json:"season"`
			StandingsLists []struct {
				Season          string `json:"season"`
				Round           string `json:"round"`
				DriverStandings []struct {
					Position     string `json:"position"`
					PositionText string `json:"positionText"`
					Points       string `json:"points"`
					Wins         string `json:"wins"`
					Driver       struct {
						DriverID        string `json:"driverId"`
						PermanentNumber string `json:"permanentNumber"`
						Code            string `json:"code"`
						URL             string `json:"url"`
						GivenName       string `json:"givenName"`
						FamilyName      string `json:"familyName"`
						DateOfBirth     string `json:"dateOfBirth"`
						Nationality     string `json:"nationality"`
					}
					Constructors []struct {
						ConstructorID string `json:"constructorId"`
						URL           string `json:"url"`
						Name          string `json:"name"`
						Nationality   string `json:"nationality"`
					}
				}
			}
		}
	}
}

// Type that represents the F1 World Constructor Championship standings.
type CStandings struct {
	MRData struct {
		XMLNS          string `json:"xmlns"`
		Series         string `json:"series"`
		URL            string `json:"url"`
		Limit          string `json:"limit"`
		Offset         string `json:"offset"`
		Total          string `json:"total"`
		StandingsTable struct {
			Season         string `json:"season"`
			StandingsLists []struct {
				Season               string `json:"season"`
				Round                string `json:"round"`
				ConstructorStandings []struct {
					Position     string `json:"position"`
					PositionText string `json:"positionText"`
					Points       string `json:"points"`
					Wins         string `json:"wins"`
					Constructor  struct {
						ConstructorID string `json:"constructorId"`
						URL           string `json:"url"`
						Name          string `json:"name"`
						Nationality   string `json:"nationality"`
					}
				}
			}
		}
	}
}

// Type that represents a quiz score.
type Score struct {
	Nick   string
	Points int
}

// Type that represents a list of scores.
// This type is needed so that we can sort the score by points (value).
// Internally score is a map[string]int, but fmt only sorts maps by key.
// We use sort.Sort() in cmdQuiz which requires ScoreList to implement the sort interface.
type ScoreList []Score

func (s ScoreList) Len() int {
	return len(s)
}

func (s ScoreList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ScoreList) Less(i, j int) bool {
	return s[i].Points < s[j].Points
}

// Small utility function that returns weather a slice of strings contains a given string.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Small utility function that reads a CSV file and returns the data as slice of slice of strings.
func readCSV(path string) (data [][]string, err error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		err = errors.New("Error opening CSV file: " + path + ".")
		return
	}
	r := csv.NewReader(f)
	data, err = r.ReadAll()
	if err != nil {
		err = errors.New("Error reading data from: " + path + ".")
		return
	}
	return
}

// Small utility function that writes a slice of slice of strings to a CSV file.
func writeCSV(path string, data [][]string) (err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		err = errors.New("Error opening CSV file: " + path + ".")
		return
	}
	w := csv.NewWriter(f)
	err = w.WriteAll(data)
	if err != nil {
		err = errors.New("Error writing data to: " + path + ".")
		return
	}
	return
}

// Small utility function that fetches and returns raw data from an URL using HTTP.
func getURL(url string) (data []byte, err error) {
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		err = errors.New("Error getting HTTP data.")
		return
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		err = errors.New("Error getting HTTP data.")
		return
	}
	return
}

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
			event = []string{e[0], e[1], e[2], e[3]}
			return event, nil
		}
	}
	err = errors.New("No event found.")
	return
}

// The tskFeeds function runs in the background as a goroutine polling a collection of news feeds.
func tskFeeds(irccon *irc.Connection) {
	var timeFormat = "2006-01-02 15:04:05 +0000 UTC" // Time format string used by the time package.
	// Loop that runs every feedInterval seconds opening the feeds CSV file and fetching news.
	for {
		time.Sleep(feedInterval * time.Second)
		feeds, err := readCSV(feedsFile)
		if err != nil {
			log.Println("tskFeeds:", err)
			continue
		}
		for key, value := range feeds {
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(value[1])
			if err != nil {
				log.Println("feed:", err)
				continue
			}
			for _, item := range feed.Items {
				// The lastTime variable keeps track of when the last feed item was retrieved.
				// If we cannot parse the time (first time) then we use timeFormat as lastTime.
				// We could use any time in the past here, but timeFormat is already available.
				lastTime, err := time.Parse(timeFormat, feeds[key][3])
				if err != nil {
					lastTime, _ = time.Parse(timeFormat, timeFormat)
				}
				itemTime := item.PublishedParsed
				// We only want to show a feed item if itemTime > lastTime.
				// Additionally we also want to make sure the feed item is no older than 2 hours.
				// This assures only current news when restarting the bot or changing the feeds.
				if itemTime.After(lastTime) && time.Since((*itemTime)) < 2*hns {
					irccon.Privmsg(feeds[key][2], fmt.Sprintf("\x02[%s] [%s]\x02", feeds[key][0], item.Title))
					irccon.Privmsg(feeds[key][2], item.Link)
					feeds[key][3] = fmt.Sprintf("%s", itemTime)
					writeCSV(feedsFile, feeds)
					time.Sleep(1 * time.Second)
				}
			}
		}
	}
}

// The tskEvents function runs in the background as a goroutine polling for new events.
func tskEvents(irccon *irc.Connection, channel string) {
	var announced [5]string                    // Small buffer to hold recently announced events.
	var index = 0                              // Index used to reference the buffer above.
	var timeFormat = "2006-01-02 15:04:05 UTC" // Time format string used by the time package.
	// This is a separate thread, we must check if the main one is connected to IRC.
	// While not connected to IRC sleep for 10 seconds before trying again.
	// If eventually a connection is established we jump out of this loop and resume.
	for !irccon.Connected() {
		log.Println("tskEvents: Waiting for an IRC connection.")
		time.Sleep(10 * time.Second)
	}
	// Loop that runs every minute opening the events CSV file and querying any event that starts within 5 minutes.
	for {
		time.Sleep(60 * time.Second)
		event, err := findNext("any", "any")
		if err != nil {
			log.Println("tskEvents:", err)
			continue
		}
		t, err := time.Parse(timeFormat, event[3])
		if err != nil {
			log.Println("tskEvents: Error parsing time.")
			continue
		}
		delta := time.Until(t)
		if delta.Minutes() > 5 {
			continue
		}
		// If the index becomes greather than what the buffer can hold, we reset it.
		// Otherwise we check if the announced buffer already contains the next event.
		// If it doesn't, the event is announced on the channel and added to the buffer.
		if index > 4 {
			index = 0
		} else {
			if !contains(announced[0:5], event[0]+" "+event[1]+" "+event[2]) {
				irccon.Privmsg(
					channel,
					fmt.Sprintf(
						"\x034Starting in 5 minutes:\x03 \x02%s %s %s\x02",
						event[0], event[1], event[2]))
				announced[index] = event[0] + " " + event[1] + " " + event[2]
				index++
			}
		}
	}
}

// The help command receives an irc connection pointer, a channel and a nick.
// It then shows a help message listing all the possible commands of the bot.
func cmdHelp(irccon *irc.Connection, channel string, nick string, search string) {
	help := [7]string{
		prefix + "bet <xxx> <yyy> <zzz> - Place a bet for the next F1 race.",
		prefix + "help - Show this help message.",
		prefix + "next [category] - Show the next motorsport event.",
		prefix + "quiz [number] - Start an F1 quiz game.",
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

// The standings command receives an irc connection pointer, a channel, a nick and a championship string.
// It then queries the Ergast F1 API for either the WDC or the WCC and displays the results on the channel.
func cmdStandings(irccon *irc.Connection, channel string, nick string, championship string) {
	var output string
	// Base URL of the Ergast F1 API.
	url := "http://ergast.com/api/f1/current/"
	if strings.ToLower(championship) == "constructor" || strings.ToLower(championship) == "constructors" {
		championship = "constructor"
		url += "constructorStandings.json"
	} else {
		championship = "driver"
		url += "driverStandings.json"
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

// The next command receives an irc connection pointer, a channel, a nick and an optional search string.
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
	event, err := findNext("[formula 1]", "race")
	if err != nil {
		irccon.Privmsg(channel, "Error finding next race.")
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

// The parsecmd function takes a message string and breaks it down into a Command.
func parseCommand(message string, nick string, channel string) (command Command, err error) {
	if len(message) > 1 && strings.HasPrefix(message, prefix) {
		split := strings.Split(message, " ")
		command.Name = split[0][1:]
		command.Args = split[1:]
		command.Nick = nick
		command.Channel = channel
		return
	} else {
		err = errors.New("parsecmd: Invalid command.")
		return
	}
}

// The quiz command receives an irc connection pointer, an irc channel and a [2]string channel.
// It runs as a goroutine that opens a quiz file and asks questions on the given irc channel.
// It then waits for answers to classify as correct or wrong or times out after a while.
func cmdQuiz(irccon *irc.Connection, channel string, c chan [2]string, number string) {
	quiz = true
	score := make(map[string]int)
	n, err := strconv.Atoi(number)
	if err != nil || (n <= 0 || n > 10) {
		n = 5
	}
	questions, err := readCSV(quizFile)
	if err != nil {
		irccon.Privmsg(channel, "Error reading questions.")
		log.Println("cmdQuiz:", err)
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

func cmdQuote(irccon *irc.Connection, channel string, nick string, args []string) {
	quotes, err := readCSV(quotesFile)
	if err != nil {
		irccon.Privmsg(channel, "Error getting quote.")
		log.Println("cmdQuote:", err)
		return
	}
	if len(args) == 0 || (len(args) > 0 && strings.ToLower(args[0]) == "get") {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(quotes))
		irccon.Privmsg(channel, fmt.Sprintf("%s - %s", quotes[index][1], quotes[index][0]))
	} else if len(args) > 1 && strings.ToLower(args[0]) == "add" {
		quotes = append(quotes, []string{time.Now().Format("02-01-2006"), strings.Join(args[1:], " ")})
		err = writeCSV(quotesFile, quotes)
		if err != nil {
			irccon.Privmsg(channel, "Error adding quote.")
			log.Println("cmdQuote:", err)
			return
		}
		irccon.Privmsg(channel, "Quote added.")
	} else {
		irccon.Privmsg(channel, "Usage: !quote [add|get] [text]")
	}
}

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
			if quiz {
				c <- [2]string{event.Nick, event.Message()}
			}
		} else {
			switch strings.ToLower(command.Name) {
			case "commands", "help":
				cmdHelp(irccon, command.Channel, command.Nick, strings.Join(command.Args, ""))
			case "next":
				cmdNext(irccon, command.Channel, command.Nick, strings.Join(command.Args, " "))
			case "bet22":
				cmdBet(irccon, command.Channel, command.Nick, command.Args)
			case "quiz":
				if !quiz {
					go cmdQuiz(irccon, command.Channel, c, strings.Join(command.Args, " "))
				}
			case "quote":
				cmdQuote(irccon, command.Channel, command.Nick, command.Args)
			case "wdc":
				go cmdStandings(irccon, command.Channel, command.Nick, "driver")
			case "wcc":
				go cmdStandings(irccon, command.Channel, command.Nick, "constructor")
			}
		}
	})
	go tskEvents(irccon, channels)
	go tskFeeds(irccon)
	irccon.Loop()
}
