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
	"encoding/csv"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Small utility function that returns weather a slice of strings contains a given string.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Small utility function that returns a map[string]string given a [][]string and a key/value pair.
func toStringMap(s [][]string, key int, value int) (result map[string]string, err error) {
	if len(s) == 0 || key < 0 || value < 0 || key >= value {
		err := errors.New("Invalid slice, key or value.")
		return result, err
	}
	result = make(map[string]string)
	for _, v := range s {
		if (key > (len(v) - 1)) || (value > (len(v) - 1)) {
			err := errors.New("Invalid slice, key or value.")
			return result, err
		}
		result[v[key]] = v[value]
	}
	return
}

// Small utility function that returns weather a nick is a user or not.
func isUser(nick string, users [][]string) bool {
	for _, user := range users {
		if strings.ToLower(nick) == strings.ToLower(user[0]) {
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
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

// Small utility function that reads messages from an input file.
func readIn(path string) (message string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		err = errors.New("Error reading message from: " + path + ".")
		return
	}
	message = string(data)
	os.Truncate(path, 0)
	return
}

// Small utility function that writes messages to an output file.
func writeOut(path string, message string) (err error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		err = errors.New("Error opening output file: " + path + ".")
		return
	}
	_, err = f.WriteString(message)
	if err != nil {
		err = errors.New("Error writing message to: " + path + ".")
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

// Small utility function that takes a message string and breaks it down into a Command.
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

// Small utility function that checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
