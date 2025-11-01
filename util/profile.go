package hlutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Profile struct {
	Avatar   string
	Domain   string
	Bio      string
	Twitter  string
	Discord  string
	Telegram string
	Website  string
}

func FetchProfilePage(profile Profile) string {
	jsonData, err := json.Marshal(profile)
	if err != nil {
		return ""
	}

	resp, err := http.Post("http://hlprofile:3000/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}
