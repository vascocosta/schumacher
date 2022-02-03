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

// Type that represents a score that can be sorted by points.
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
