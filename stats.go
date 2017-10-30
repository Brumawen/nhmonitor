package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Stats holds the NiceHash statistics retrieved from the web service.
type Stats struct {
	Result struct {
		Stats []struct {
			Balance       string `json:"balance"`
			RejectedSpeed string `json:"rejected_speed"`
			Algo          int    `json:"algo"`
			AcceptedSpeed string `json:"accepted_speed"`
		}
		Addr string `json:"addr"`
	}
	Method string
}

// GetStats gets the NiceHash statistics from the web service.
func GetStats(add string) (Stats, error) {
	url := "https://api.nicehash.com/api?method=stats.provider&addr="
	response, err := http.Get(fmt.Sprintf("%s%s", url, add))
	if err != nil {
		log.Println("Error getting Stats.", err)
		return Stats{}, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading Stats response body.", err)
		return Stats{}, err
	}
	s, err := getStatsFromJSON(data)
	if err != nil {
		t := string(data)
		if strings.Contains(t, "<!DOCTYPE html>") {
			log.Println("NiceHash web service is offline.")
		} else {
			log.Println("Error deserializing stats data.", err)
		}
		return Stats{}, err
	}
	return s, nil
}

// GetBalance returns the total outstanding balance for this wallet
func (s *Stats) GetBalance() float64 {
	var r float64
	for _, s := range s.Result.Stats {
		b, err := strconv.ParseFloat(s.Balance, 64)
		if err != nil {
			log.Println("Failed to parse balance", s, err)
		} else {
			r += b
		}
	}
	return r
}

// getStatsFromJSON deserializes the json string into a Stats struct
func getStatsFromJSON(j []byte) (Stats, error) {
	var s Stats
	err := json.Unmarshal(j, &s)
	if err != nil {
		return Stats{}, err
	}
	return s, nil
}
