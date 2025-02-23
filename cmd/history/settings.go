package main

import (
	"advisordev/internal/candles/update"
	"encoding/xml"
	"os"
)

type Settings struct {
	SecurityCodes []update.SecurityCode `xml:"SecurityCodes>SecurityCode"`
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
