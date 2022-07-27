package userendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/queries"
)

type Auth struct {
	DeviceID      string `json:"deviceId"`
	DeviceName    string `json:"deviceName"`
	UserAgent     string `json:"User-Agent"`
	ApVersion     string `json:"Apversion"`
	Authorization string `json:"authorization"`
}

func Get(w http.ResponseWriter, r *http.Request) {
	users := queries.GetArray(r.URL.Query(), "users")
	if len(users) == 0 {
		http.Error(w, "No users specified", http.StatusBadRequest)
		return
	}

	output, err := queries.GetUsers(context.Background(), users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func Post(w http.ResponseWriter, r *http.Request) {
	var auth Auth
	err := json.NewDecoder(r.Body).Decode(&auth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api-flmg.eggs.mu/v1/users/users/profile", nil)
	req.Header.Set("Authorization", auth.Authorization)
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("apversion", auth.ApVersion)
	req.Header.Set("deviceid", auth.DeviceID)
	req.Header.Set("devicename", auth.DeviceName)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://eggs.mu")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user queries.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user.Data.EggsID == "" {
		http.Error(w, "Invalid token", http.StatusInternalServerError)
		return
	}

	eggsID, token, err := queries.GetUserCredentials(context.Background(), user)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if eggsID == "" {
		token, err = generateToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = queries.InsertUser(context.Background(), user, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if token == "" {
		token, err = generateToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = queries.UpdateUserToken(context.Background(), user, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = queries.UpdateUserDetails(context.Background(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, token)
}

func generateToken() (token string, err error) {
	token, err = queries.GenerateRandomString(64)
	return
}
