package main

import (
	"strings"
	"os"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var baseUrl string = "http://www.omdbapi.com/"

type SingleResultResponse struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	Rated      string `json:"Rated"`
	Released   string `json:"Released"`
	Runtime    string `json:"Runtime"`
	Genre      string `json:"Genre"`
	Director   string `json:"Director"`
	Writer     string `json:"Writer"`
	Actors     string `json:"Actors"`
	Plot       string `json:"Plot"`
	Language   string `json:"Language"`
	Country    string `json:"Country"`
	Awards     string `json:"Awards"`
	Poster     string `json:"Poster"`
	Metascore  string `json:"Metascore"`
	ImdbRating string `json:"imdbRating"`
	ImdbVotes  string `json:"imdbVotes"`
	ImdbId     string `json:"imdbId"`
	Type       string `json:"Type"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}

func SearchByTitle(searchTerm string) (SingleResultResponse, error) {
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return SingleResultResponse{}, err
	}

	query := req.URL.Query()
	query.Add("t", searchTerm)
	query.Add("apikey", "70a330ff")
	req.URL.RawQuery = query.Encode()

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return SingleResultResponse{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SingleResultResponse{}, err
	}

	var response SingleResultResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return SingleResultResponse{}, err
	}

	return response, nil
}

func main() {
	result, err := SearchByTitle(strings.Join(os.Args[2:], " "))
	if err != nil || result.Title == "" {
		fmt.Println("Cannot find movie or show.")
		return
	}
	fmt.Println(fmt.Sprintf("Title: %s | Year: %s | Genre: %s | Director: %s | IMDB Rating: %s\n\n%s",
				result.Title,
				result.Year,
				result.Genre,
				result.Director,
				result.ImdbRating,
				result.Plot))
}
