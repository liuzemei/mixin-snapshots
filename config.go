package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type cfg struct {
	Dev      string `json:"dev"`
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Name     string `json:"name"`
	} `json:"database"`
	Mixin struct {
		ClientID   string `json:"client_id"`
		SessionID  string `json:"session_id"`
		PrivateKey string `json:"private_key"`
	}
}

var config cfg

func init() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println("config.json open fail...", err)
		return
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Println("config.json parse err...", err)
	}
}
