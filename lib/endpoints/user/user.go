package userendpoint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

type Auth struct {
	DeviceID      string `json:"deviceId"`
	DeviceName    string `json:"deviceName"`
	UserAgent     string `json:"User-Agent"`
	ApVersion     string `json:"Apversion"`
	Authorization string `json:"authorization"`
}

func Get(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	users := queries.GetArray(r.URL.Query(), "users")
	if len(users) == 0 {
		return logging.SE(http.StatusBadRequest, errors.New("no users specified"))
	}

	output, err := queries.GetUsers(context.Background(), users)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}

	b, err := json.Marshal(output)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	w.Write(b)
	return nil
}

func Post(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var auth Auth
	err := json.Unmarshal(b, &auth)
	if err != nil {
		return logging.SE(http.StatusBadRequest, err)
	}

	client := &http.Client{Timeout: 1 * time.Minute}
	req, err := http.NewRequest("GET", "https://api-flmg.eggs.mu/v1/users/users/profile", nil)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", auth.Authorization)
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("apversion", auth.ApVersion)
	req.Header.Set("deviceid", auth.DeviceID)
	req.Header.Set("devicename", auth.DeviceName)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://eggs.mu")
	resp, err := client.Do(req)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return logging.SE(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	var user queries.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	if user.Data.EggsID == "" {
		return logging.SE(http.StatusBadRequest, errors.New("invalid token"))
	}

	eggsID, token, err := queries.GetUserCredentials(context.Background(), user)
	if err != nil && err.Error() != "no rows in result set" {
		return logging.SE(http.StatusInternalServerError, err)
	}

	if eggsID == "" {
		token, err = generateToken()
		if err != nil {
			return logging.SE(http.StatusInternalServerError, err)
		}
		err = queries.InsertUser(context.Background(), user, token)
		if err != nil {
			return logging.SE(http.StatusInternalServerError, err)
		}
	}

	if token == "" {
		token, err = generateToken()
		if err != nil {
			return logging.SE(http.StatusInternalServerError, err)
		}
		err = queries.UpdateUserToken(context.Background(), user, token)
		if err != nil {
			return logging.SE(http.StatusInternalServerError, err)
		}
	}

	err = queries.UpdateUserDetails(context.Background(), user)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}

	fmt.Fprint(w, token)
	return nil
}

func generateToken() (token string, err error) {
	token, err = queries.GenerateRandomString(64)
	return
}
