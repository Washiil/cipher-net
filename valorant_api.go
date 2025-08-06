package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var httpClient = &http.Client{}

type AccountResponse struct {
	Data struct {
		PUUID string `json:"puuid"`
	} `json:"data"`
}

func NameToUUID(name string, tag string, token string) (string, error) {
	url := fmt.Sprintf("https://api.henrikdev.xyz/valorant/v1/account/%s/%s", name, tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed AccountResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}

	return parsed.Data.PUUID, nil
}
