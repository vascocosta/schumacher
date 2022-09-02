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

const (
	server        = "irc.quakenet.org:6667"                // Hostname of the server to connect to.
	prefix        = "!"                                    // Prefix which is used by the user to issue commands.
	folder        = "/home/gluon/var/irc/bots/Schumacher/" // Full path to the folder.
	answersFile   = folder + "answers.csv"                 // Full path to the answers file.
	betsFile      = folder + "bets.csv"                    // Full path to the bets file.
	driversFile   = folder + "drivers.csv"                 // Full path to the drivers file.
	eventsFile    = folder + "events.csv"                  // Full path to the events file.
	feedsFile     = folder + "feeds.csv"                   // Full path to the feeds file.
	usersFile     = folder + "users.csv"                   // Full path to the users file.
	resultsFile   = folder + "results.csv"                 // Full path to the results file.
	quizFile      = folder + "quiz.csv"                    // Full path to the quiz file.
	quotesFile    = folder + "quotes.csv"                  // Full path to the quotes file.
	weatherFile   = folder + "weather.csv"                 // Full path to the weather file.
	pluginsFolder = folder + "plugins/"                    // Full path to the plugins folder.
	pollTimeout   = 60                                     // Poll timeout in seconds.
	quizTimeout   = 20                                     // Quiz timeout in seconds.
	hns           = 3600000000000                          // Number of nanoseconds in one hour.
	feedInterval  = 300                                    // Feed poll interval in seconds.
	owmAPIKey     = "f97b1089707bd013b60c22db86730cf8"     // OpenWeatherMap API key.
)
