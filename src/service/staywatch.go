package service

import (
	"encoding/json"
	"io"
	"net/http"
)

func GetStayWatchMember() ([]StaywatchUsers, error) {
	var users []StaywatchUsers
	req, err := http.NewRequest("GET", staywatch.BaseURL+staywatch.Users, nil)
	if err != nil {
		return nil, err
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}

	return users, nil
}
