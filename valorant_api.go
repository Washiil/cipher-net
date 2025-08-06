package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var httpClient = &http.Client{}

type AccountResponse struct {
	Status int `json:"status"`
	Data   struct {
		PUUID        string `json:"puuid"`
		Region       string `json:"region"`
		AccountLevel int    `json:"account_level"`
		Name         string `json:"name"`
		Tag          string `json:"tag"`
		Card         struct {
			Small string `json:"small"`
			Large string `json:"large"`
			Wide  string `json:"wide"`
			ID    string `json:"id"`
		} `json:"card"`
		LastUpdate    string `json:"last_update"`
		LastUpdateRaw int64  `json:"last_update_raw"`
	} `json:"data"`
}

func NameToUUID(name, tag, token string) (string, error) {
	url := fmt.Sprintf("https://api.henrikdev.xyz/valorant/v1/account/%s/%s", name, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s (status %d)", string(body), resp.StatusCode)
	}

	var result AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Data.PUUID, nil
}
