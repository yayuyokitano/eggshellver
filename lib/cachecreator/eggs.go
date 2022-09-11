package cachecreator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

const baseurl = "https://api-flmg.eggs.mu/v1"

func eggsGet(url string, v any) (err error) {
	client := &http.Client{Timeout: 1 * time.Minute}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", baseurl, url), nil)
	if err != nil {
		return
	}
	t := time.Now()

	req.Header.Set("Authorization", os.Getenv("TESTUSER_AUTHORIZATION"))
	req.Header.Set("User-Agent", os.Getenv("TESTUSER_USERAGENT"))
	req.Header.Set("apversion", os.Getenv("TESTUSER_APVERSION"))
	req.Header.Set("deviceid", os.Getenv("TESTUSER_DEVICEID"))
	req.Header.Set("devicename", os.Getenv("TESTUSER_DEVICENAME"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://eggs.mu")

	logging.LogFetch(req)
	resp, err := client.Do(req)
	if err != nil {
		logging.LogFetchErrored(req, resp)
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.LogFetchErrored(req, resp)
		return
	}
	err = json.Unmarshal(respBody, v)
	if err != nil {
		return
	}
	logging.LogFetchCompleted(resp, respBody, t)
	return
}

func getSongs(offset int, limit int) (songs queries.SearchSongResp) {
	err := eggsGet(fmt.Sprintf("search/search/musics?musicTitle=%%25&offset=%d&limit=%d", offset, limit), &songs)
	if err != nil {
		panic(err)
	}
	return
}

func getRecentSongs() (songs queries.SearchSongResp, err error) {
	err = eggsGet("artists/new/musics?limit=100", &songs)
	return
}
