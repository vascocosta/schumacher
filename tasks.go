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
	"fmt"
	"github.com/gocolly/colly"
	"github.com/mmcdole/gofeed"
	"github.com/thoj/go-ircevent"
	"log"
	"mvdan.cc/xurls/v2"
	"strings"
	"time"
)

// The tskFeeds function runs in the background as a goroutine polling a collection of news feeds.
func tskFeeds(irccon *irc.Connection) {
	var timeFormat = "2006-01-02 15:04:05 +0000 UTC" // Time format string used by the time package.
	// Loop that runs every feedInterval seconds opening the feeds CSV file and fetching news.
	for {
		time.Sleep(feedInterval * time.Second)
		start := time.Now()
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
		fmt.Printf("Feed processing time: %s", time.Since(start))
	}
}

// The tskFeeds2 function runs in the background as a goroutine polling a collection of news feeds.
func tskFeeds2(irccon *irc.Connection) {
	// Simple structure type used to send feed data to a go channel.
	// It stores a key that indexes each different feed and a value.
	// This allows the reading thread (this function) to access those two variables from the channel.
	// The key is required so that the reading thread can update the lastTime field of each feed.
	type FeedData struct {
		Key   int
		Value *gofeed.Feed
	}
	var timeFormat = "2006-01-02 15:04:05 +0000 UTC" // Time format string used by the time package.
	// Loop that runs every feedInterval seconds opening the feeds CSV file and fetching news.
	for {
		time.Sleep(feedInterval * time.Second)
		//start := time.Now()
		feeds, err := readCSV(feedsFile)
		feedDataCh := make(chan FeedData)
		if err != nil {
			log.Println("tskFeeds:", err)
			continue
		}
		// Loop that spawns a goroutine worker thread per each feed source in the feeds CSV file.
		// The annonymous goroutine function accepts the k and v parameters, passed as arguments.
		// This is to avoid undesired indeterministic effects from using a closure as a goroutine.
		// The goroutine builds a Feed type by parsing the URL field for each feed in the CSV file.
		// A FeedData type is built and sent to the go channel to be received by the reading thread.
		for key, value := range feeds {
			go func(k int, v []string) {
				fp := gofeed.NewParser()
				feed, err := fp.ParseURL(v[1])
				if err != nil {
					log.Println("feed:", err)
					return
				}
				feedData := FeedData{k, feed}
				feedDataCh <- feedData
			}(key, value)
		}
		// Loop that runs a select on the go channel for as long as there's data to be read or until a timeout occurs.
		// In case feedData can be read from the communication channel, process all the feed items and show new ones.
		// In case this thread needs to wait more than 2 minutes to receive data from the goroutines a tiemout occurs.
		for {
			timeout := false
			select {
			case feedData := <-feedDataCh:
				for _, item := range feedData.Value.Items {
					// The lastTime variable keeps track of when the last feed item was retrieved.
					// If we cannot parse the time (first time) then we use timeFormat as lastTime.
					// We could use any time in the past here, but timeFormat is already available.
					lastTime, err := time.Parse(timeFormat, feeds[feedData.Key][3])
					if err != nil {
						lastTime, _ = time.Parse(timeFormat, timeFormat)
					}
					itemTime := item.PublishedParsed
					// We only want to show a feed item if itemTime > lastTime.
					// Additionally we also want to make sure the feed item is no older than 2 hours.
					// This assures only current news when restarting the bot or changing the feeds.
					if itemTime.After(lastTime) && time.Since((*itemTime)) < 2*hns {
						irccon.Privmsg(
							feeds[feedData.Key][2],
							fmt.Sprintf("\x02[%s] [%s]\x02", feeds[feedData.Key][0], item.Title))
						irccon.Privmsg(feeds[feedData.Key][2], item.Link)
						feeds[feedData.Key][3] = fmt.Sprintf("%s", itemTime)
						writeCSV(feedsFile, feeds)
						time.Sleep(1 * time.Second)
					}
				}
			case <-time.After(60 * time.Second):
				timeout = true
				break // Break out of the select statement.
			}
			if timeout {
				break // We need this second break when a timeout occurs to break out of the select loop.
			}
		}
		//fmt.Printf("Feed processing time: %s\n", time.Since(start)-2*time.Minute)
	}
}

// The tskEvents function runs in the background as a goroutine polling for new events.
func tskEvents(irccon *irc.Connection) {
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
					event[4],
					fmt.Sprintf(
						"\x034Starting in 5 minutes:\x03 \x02%s %s %s\x02",
						event[0], event[1], event[2]))
				announced[index] = event[0] + " " + event[1] + " " + event[2]
				index++
				users, err := readCSV(usersFile)
				if err != nil {
					log.Println("tksEvents:", err)
					continue
				}
				var mentions string
				for _, user := range users {
					channels := strings.Split(user[3], ":")
					if contains(channels, event[4]) {
						mentions += user[0] + " "
					}
				}
				if mentions != "" {
					irccon.Privmsg(event[4], mentions)
					irccon.Privmsg(event[4], "Use !notify off to stop getting mentions for events on this channel.")
				}
			}
		}
	}
}

// The tskHTMLTitle function runs in the background as a goroutine that extracts HTML titles from links.
func tskHTMLTitle(irccon *irc.Connection, channel string, message string) {
	rxStrict := xurls.Strict()
	url := rxStrict.FindString(message)
	c := colly.NewCollector()
	c.OnHTML("title", func(e *colly.HTMLElement) {
		irccon.Privmsg(channel, "Title: "+e.Text)
	})
	c.Visit(url)
}
