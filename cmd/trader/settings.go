package main

import (
	"advisordev/internal/trader"
	"encoding/xml"
	"os"
)

type Settings struct {
	AdvisorUrl      string
	Clients         []trader.Client         `xml:"Clients>Client"`
	StrategyConfigs []trader.StrategyConfig `xml:"StrategyConfigs>StrategyConfig"`
}

func loadSettings(filePath string) (Settings, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Settings{}, err
	}
	defer file.Close()

	var result Settings
	err = xml.NewDecoder(file).Decode(&result)
	if err != nil {
		return Settings{}, err
	}
	return result, nil
}
