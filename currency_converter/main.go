package main

import (
	"encoding/json"
	"net/http"

	//"github.com/labstack/echo/v5"
	"log"
)

func main() {
	url := "https://frankfurter.dev"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Connection err: %v", err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		log.Fatalf("Incorrect status: %v", resp.Status)
	}

	var rates []RateData
	err = json.NewDecoder(resp.Body).Decode(&rates)
	if err != nil {
		log.Fatalf("Error with decoding: %v", err)
	}
}

// Struct

type RateData struct {
	Data  string  `json:"data"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}
