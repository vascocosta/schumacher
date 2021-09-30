package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/thoj/go-ircevent"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const nick = "Schumacher2"                                            // Nick to be used by the bot.
const channels = "#motorsport"                                        // Names of the channels to join.
const server = "irc.quakenet.org:6667"                                // Hostname of the server to connect to.
const prefix = "!"                                                    // Prefix which is used by the user to issue commands.
const dbPath = "/home/gluon/var/irc/bots/Senna/data/Motorsport.db"    // Full path to the database.
const betsFile = "/home/gluon/var/irc/bots/Schumacher/bets.csv"       // Full path to the bets file.
const driversFile = "/home/gluon/var/irc/bots/Schumacher/drivers.csv" // Full path to the drivers file.
const eventsFile = "/home/gluon/var/irc/bots/Schumacher/events.csv"   // Full path to the events file.
const quizFile = "/home/gluon/var/irc/bots/Schumacher/quiz.csv"       // Full path to the quiz file.

var quiz bool

// Type that represents an IRC command issued by the user.
type Command struct {
	Name    string
	Args    []string
	Nick    string
	Channel string
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

// Small utility function that parses a CSV file and returns the data as slice of slice of strings.
func csvToSlice(path string) (data [][]string, err error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		err = errors.New("Error opening CSV file.")
		return
	}
	r := csv.NewReader(f)
	data, err = r.ReadAll()
	if err != nil {
		err = errors.New("Error reading data.")
		return
	}
	return
}

// The announce function runs in the background as a goroutine polling for new events.
func announce(irccon *irc.Connection, channel string) {
	var announced [5]string // Small buffer to hold recently announced events.
	var index = 0           // Index used to reference the buffer above.
	var dtstart string      // Event start time.
	var summary string      // Event description.
	for !irccon.Connected() {
		log.Println("announce: Waiting for an IRC connection.")
		time.Sleep(10 * time.Second)
	}
	// Loop that runs every minute opening the database and querying any event that starts within 5 minutes.
	for {
		time.Sleep(60 * time.Second)
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Println("announce: Error opening database.")
			db.Close()
			continue
		}
		row := db.QueryRow("SELECT dtstart, summary FROM events WHERE dtstart > datetime(\"now\") " +
			"AND dtstart < datetime(\"now\", \"+5 Minute\") LIMIT 1")
		err = row.Scan(&dtstart, &summary)
		if err != nil {
			if err.Error() != "sql: no rows in result set" {
				log.Println("announce: Error querying database.")
			}
			db.Close()
			continue
		}
		// If the index becomes greather than what the buffer can hold, we reset it.
		if index > 4 {
			index = 0
		} else {
			if !contains(announced[0:4], summary) {
				irccon.Privmsg(channel, "\x034Starting in 5 minutes:\x03 "+summary)
				announced[index] = summary
				index++
			}
		}
		db.Close()
	}
}

// The next command receives an irc connection pointer, a channel, a nick and an optional search string.
// It then queries the database and returns which event(s) are happening next, showing them on the channel.
func cmdNext(irccon *irc.Connection, channel string, nick string, search string) {
	var tz string      // Time zone of each user.
	var dtstart string // Event start time.
	var summary string // Event description.
	// Open database and query the user's time zone.
	db, err := sql.Open("sqlite3", dbPath)
	defer db.Close()
	if err != nil {
		irccon.Privmsg(channel, "Error opening database.")
		log.Println("cmdNext: Error opening database.")
		return

	}
	row := db.QueryRow("SELECT tz FROM users WHERE nick = ? LIMIT 1", nick)
	err = row.Scan(&tz)
	if err != nil {
		tz = "Europe/Berlin"
	}
	// Do some search string replacements in case there's actually a search argument.
	// Users use abreviated search terms, which are expanded for better database matching.
	// Else, simply retrieve the next event that is happening closest to the current time.
	if search != "" {
		switch search {
		case "f1":
			search = "Formula 1"
		case "f2":
			search = "Formula 2"
		case "f3":
			search = "Formula 3"
		case "qualy":
			search = "Formula 1%Qualifying"
		case "race":
			search = "Formula 1%Race"
		}
		row = db.QueryRow("SELECT dtstart, summary FROM events WHERE dtstart > datetime(\"now\") "+
			"AND summary LIKE ? ORDER BY dtstart LIMIT 1", fmt.Sprintf("%%%s%%", search))
	} else {
		row = db.QueryRow("SELECT dtstart, summary FROM events WHERE dtstart > datetime(\"now\") " +
			"ORDER BY dtstart LIMIT 1")
	}
	// Get the results from the row, do some formatting, calculate time delta and finally show the results.
	// The times are localised as per the user's time zone before being shown.
	// The time delta between now and the next event uses modulo to perfectly round days, hour an minutes.
	err = row.Scan(&dtstart, &summary)
	if err != nil {
		irccon.Privmsg(channel, "Error querying database.")
		log.Println("cmdNext: Error querying database.")
		return
	}
	t, err := time.Parse("2006-01-02 15:04:05 UTC", dtstart)
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
		wday, mday, month, hour, min, zone, uoffset, summary, days, hours, minutes))
}

func cmdBet(irccon *irc.Connection, channel string, nick string, bet []string) {
	var correct int
	var bets [][]string
	var update bool
	race, err := findNext("formula 1", "race")
	if err != nil {
		irccon.Privmsg(channel, "Error finding next race.")
		log.Println("cmdBet:", err)
		return
	}
	if len(bet) == 0 {
		f, err := os.Open(betsFile)
		defer f.Close()
		if err != nil {
			log.Println("cmdBet: Error opening file " + betsFile)
			return
		}
		r := csv.NewReader(f)
		bets, err = r.ReadAll()
		f.Close()
		if err != nil {
			log.Println("cmdBet: Error reading bets.")
			return
		}
		for i := len(bets) - 1; i >= 0; i-- {
			if strings.ToLower(bets[i][0]) == strings.ToLower(race) && strings.ToLower(bets[i][1]) == strings.ToLower(nick) {
				first := strings.ToUpper(bets[i][2])
				second := strings.ToUpper(bets[i][3])
				third := strings.ToUpper(bets[i][4])
				irccon.Privmsg(channel, fmt.Sprintf("Your current bet for the %s: %s %s %s", race, first, second, third))
				return
			}
		}
		irccon.Privmsg(channel, fmt.Sprintf("You haven't placed a bet for the %s yet.", race))
		return
	}
	if len(bet) != 3 {
		irccon.Privmsg(channel, "The bet must contain 3 drivers.")
		return
	}
	f, err := os.Open(driversFile)
	defer f.Close()
	if err != nil {
		log.Println("cmdBet: Error opening file " + driversFile)
	}
	r := csv.NewReader(f)
	drivers, err := r.ReadAll()
	f.Close()
	if err != nil {
		log.Println("cmdBet: Error reading drivers")
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
	f, err = os.Open(betsFile)
	defer f.Close()
	if err != nil {
		log.Println("cmdBet: Error opening file " + betsFile)
		return
	}
	r = csv.NewReader(f)
	bets, err = r.ReadAll()
	f.Close()
	if err != nil {
		log.Println("cmdBet: Error reading bets.")
		return
	}
	for i := 0; i < len(bets); i++ {
		if strings.ToLower(bets[i][0]) == strings.ToLower(race) && strings.ToLower(bets[i][1]) == strings.ToLower(nick) {
			update = true
			bets[i] = []string{strings.ToLower(race), strings.ToLower(nick), first, second, third, "0"}
			break
		}
	}
	if !update {
		bets = append(bets, []string{strings.ToLower(race), strings.ToLower(nick), first, second, third, "0"})
	}
	f, err = os.OpenFile(betsFile, os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		log.Println("cmdBet: Error opening file " + betsFile)
		return
	}
	w := csv.NewWriter(f)
	err = w.WriteAll(bets)
	if err != nil {
		log.Println("cmdBet: Error writing bets.", err)
		return
	}
	irccon.Privmsg(channel, "Your bet for the "+race+" was successfully updated.")
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

// The quiz command receives an irc connection pointer, an irc channel and a string channel.
// It runs as a goroutine that opens a quiz file and asks questions on the given irc channel.
// It then waits for answers to classify as correct or wrong or times out after a while.
func cmdQuiz(irccon *irc.Connection, channel string, c chan string) {
	quiz = true
	questions, err := csvToSlice(quizFile)
	if err != nil {
		log.Println("cmdQuiz:", err)
		irccon.Privmsg(channel, "Error reading questions.")
		return
	}
	// This is an obfuscated way to randomise a slice that I googled in order to randomise the questions.
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(questions), func(i, j int) { questions[i], questions[j] = questions[j], questions[i] })
	// This is the main loop of the goroutine, which for now only asks 5 questions to avoid SPAM.
	// After showing the question on irc, it waits for an answer on the "c" go channel and classifies it.
	// The answer is sent to the "c" go channel on the main goroutine, inside the "PRIVMSG" callback.
	// Eventually if no answer is sent, it times out after 15 seconds.
	for i := 0; i < 5; i++ {
		irccon.Privmsg(channel, strconv.Itoa(i+1)+"/5 - "+questions[i][0])
		select {
		case answer := <-c:
			if strings.ToLower(answer) == strings.ToLower(questions[i][1]) {
				irccon.Privmsg(channel, "Correct!")
			} else {
				if i >= 0 {
					i-- // Avoid advancing to the next question, when answer is wrong.
				}
				irccon.Privmsg(channel, "Wrong!")
			}
		case <-time.After(15 * time.Second):
			irccon.Privmsg(channel, "Time's up... The correct answer was: "+questions[i][1])
		}
	}
	quiz = false
	irccon.Privmsg(channel, "The quiz is over!")
}

func findNext(category string, session string) (eventName string, err error) {
	events, err := csvToSlice(eventsFile)
	if err != nil {
		return
	}
	for _, event := range events {
		if strings.ToLower(event[0]) == strings.ToLower(category) && strings.ToLower(event[2]) == strings.ToLower(session) {
			t, err := time.Parse("2006-01-02 15:04:05 UTC", event[3])
				if err != nil {
					log.Println("findNext: Error parsing time.")
					return "", err
				}
			delta := time.Until(t)
			if delta >= 0 {
				eventName = event[1]
				return eventName, nil
			}
		}
	}
	err = errors.New("No event found.")
	return
}

func main() {
	c := make(chan string)
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
				c <- m
			}
		} else {
			switch strings.ToLower(command.Name) {
			case "next":
				cmdNext(irccon, command.Channel, command.Nick, strings.Join(command.Args, " "))
			case "bet22":
				// t1 := time.Now()
				cmdBet(irccon, command.Channel, command.Nick, command.Args)
				// t2 := time.Now()
				// duration := t2.Sub(t1)
				// log.Println(duration.Microseconds())
			case "quiz":
				if !quiz {
					go cmdQuiz(irccon, command.Channel, c)
				}
			}
		}
	})
	go announce(irccon, "#motorsport")
	irccon.Loop()
}
