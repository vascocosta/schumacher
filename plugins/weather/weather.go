package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	owm "github.com/briandowns/openweathermap"
	"log"
	"os"
	"strings"
)

const (
	folder      = "/home/gluon/var/irc/bots/Schumacher/" // Full path to the folder.
	weatherFile = folder + "weather.csv"                 // Full path to the weather file.
	owmAPIKey   = "f97b1089707bd013b60c22db86730cf8"     // OpenWeatherMap API key.
)

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

// The weather command receives an IRC connection pointer, a channel, a nick and an arguments slice of strings.
// It then shows the current weather for a given location on the channel using the OpenWeatherMap API.
func cmdWeather(nick string, args []string) {
	weather, err := readCSV(weatherFile)
	if err != nil {
		fmt.Println("Error getting weather settings.")
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
			fmt.Println("Get the weather for some location before setting the units.")
			return
		}
		err = writeCSV(weatherFile, weather)
		if err != nil {
			fmt.Println("Error storing weather units.")
			log.Println("cmdWeather:", err)
			return
		}
		fmt.Println("Temperature units updated.")
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
			fmt.Println("Error storing weather location.")
			log.Println("cmdWeather:", err)
			return
		}
	}
	if location == "" {
		fmt.Println("Please provide a location as argument.")
		return
	}
	if tempUnits == "F" {
		windUnits = "mph"
	}
	// Finally we get the current weather at a location using the temperature units.
	// Then we display a nicely formatted and compact weather string on the channel.
	w, err := owm.NewCurrent(tempUnits, "en", owmAPIKey)
	if err != nil {
		fmt.Println("Error fetching weather.")
		log.Println("cmdWeather:", err)
		return
	}
	err = w.CurrentByName(location)
	if err != nil {
		fmt.Println("Could not fetch weather for that location.")
		log.Println("cmdWeather:", err)
		return
	}
	fmt.Println(
		fmt.Sprintf("%s: %s | Temperature: %0.1f%s | Humidity: %d%% | Pressure: %0.1fhPa | Wind: %0.1f%s",
			w.Name,
			w.Weather[0].Description,
			w.Main.Temp,
			tempUnits,
			w.Main.Humidity,
			w.Main.Pressure,
			w.Wind.Speed,
			windUnits))
	/*
	        f, err := owm.NewForecast("5", tempUnits, "en", owmAPIKey)
	        if err != nil {
			log.Println(err)
	                return
	        }
	        f.DailyByName(location, 5)
	        forecast := f.ForecastWeatherJson.(*owm.Forecast5WeatherData)
	        fmt.Println(forecast)
	*/

}

func main() {
	cmdWeather(os.Args[1], os.Args[2:])
}
