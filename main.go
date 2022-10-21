package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type player struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Age   string   `json:"age"`
	Teams []string `json:"team"`
}

type data struct {
	Data struct {
		Team struct {
			ID      uint   `json:"id"`
			Name    string `json:"name"`
			Players []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Age  string `json:"age"`
			} `json:"players"`
		} `json:"team"`
	} `json:"data"`
}

func containsInArray(teams []string, str string) bool {
	for _, v := range teams {
		if v == str {
			return true
		}
	}
	return false
}

func getData(url string) data {
	var temp interface{}
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	s := data{}
	json.Unmarshal(body, &temp)
	jsonString, _ := json.Marshal(temp)
	json.Unmarshal(jsonString, &s)
	return s
}

func sequentialGetData(savedTeams *[]data) {
	var teams = []string{"Germany", "England", "France", "Spain", "Manchester United", "Arsenal", "Chelsea", "Barcelona", "Real Madrid", "Bayern Munich"}
	z := len(teams)
	for i := 1; z != 0; i++ {
		url := fmt.Sprintf("https://api-origin.onefootball.com/score-one-proxy/api/teams/en/%d.json", i)
		s := getData(url)
		c := containsInArray(teams, s.Data.Team.Name)
		if c {
			z--
			*savedTeams = append(*savedTeams, s)
		}
		println(i)
	}
}

func worker(numberOfTeams *int, teams *[]string, id <-chan int, d chan<- data) {
	for j := range id {
		url := fmt.Sprintf("https://api-origin.onefootball.com/score-one-proxy/api/teams/en/%d.json", j)
		team := getData(url)
		if containsInArray(*teams, team.Data.Team.Name) {
			d <- team
			*numberOfTeams--
		}
	}
}

func concurrentGetData(savedTeams *[]data) {
	id := make(chan int)
	d := make(chan data, 10)
	var teams = []string{"Germany", "England", "France", "Spain", "Manchester United", "Arsenal", "Chelsea", "Barcelona", "Real Madrid", "Bayern Munich"}
	numberOfTeams := len(teams)
	for j := 0; j < 50; j++ {
		go func() {
			worker(&numberOfTeams, &teams, id, d)
		}()
	}
	for i := 1; numberOfTeams != 0; i++ {
		id <- i
	}
	close(id)
	for i := 0; i < 10; i++ {
		*savedTeams = append(*savedTeams, <-d)
	}
}

func savePlayers(savedTeams *[]data, savedPlayers *map[int]player) {
	for _, s := range *savedTeams {
		for i := 0; i < len(s.Data.Team.Players); i++ {
			id, _ := strconv.Atoi(s.Data.Team.Players[i].ID)
			name := s.Data.Team.Players[i].Name
			age := s.Data.Team.Players[i].Age
			tempPlayer := player{ID: id, Name: name, Age: age, Teams: []string{s.Data.Team.Name}}
			if (*savedPlayers)[id].ID == tempPlayer.ID {
				var tempTeams []string
				tempTeams = append(tempTeams, (*savedPlayers)[id].Teams[0])
				tempTeams = append(tempTeams, s.Data.Team.Name)
				(*savedPlayers)[id] = player{id, name, age, tempTeams}
			} else {
				(*savedPlayers)[id] = tempPlayer
			}
		}
	}
}

func main() {
	startTime := time.Now()
	var savedTeams []data
	savedPlayers := make(map[int]player)
	// sequentialGetData(&savedTeams)
	concurrentGetData(&savedTeams)
	savePlayers(&savedTeams, &savedPlayers)
	keys := make([]int, 0, len(savedPlayers))
	for k := range savedPlayers {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for p, k := range keys {
		fmt.Println(p, savedPlayers[k])
	}
	log.Println("total time:", time.Since(startTime))
}
