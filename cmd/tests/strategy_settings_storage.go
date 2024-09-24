package main

import (
	"encoding/xml"
	"os"
)

type StrategySettings struct {
	SecurityCodes []SecurityCode `xml:"SecurityCodes>SecurityCode"`
}

type SecurityCode struct {
	Code      string `xml:",attr"`
	FinamCode string `xml:",attr"`
	MfdCode   string `xml:",attr"`
}

func loadStrategySettings(filePath string) (StrategySettings, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return StrategySettings{}, err
	}
	defer file.Close()

	var result StrategySettings
	err = xml.NewDecoder(file).Decode(&result)
	if err != nil {
		return StrategySettings{}, err
	}
	return result, nil
}
