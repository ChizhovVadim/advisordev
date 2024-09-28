package main

import (
	"advisordev/internal/advisors"
	"advisordev/internal/trader"
	"encoding/xml"
	"os"
)

type TraderSettings struct {
	Clients         []trader.Client           `xml:"Clients>Client"`
	StrategyConfigs []advisors.StrategyConfig `xml:"StrategyConfigs>StrategyConfig"`
}

func loadSettings(filePath string) (TraderSettings, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return TraderSettings{}, err
	}
	defer file.Close()

	var result TraderSettings
	err = xml.NewDecoder(file).Decode(&result)
	if err != nil {
		return TraderSettings{}, err
	}
	return result, nil
}
