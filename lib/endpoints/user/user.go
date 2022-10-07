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
	"github.com/yayuyokitano/eggshellver/lib/router"
)

type Auth struct {
	DeviceID      string `json:"deviceId"`
	DeviceName    string `json:"deviceName"`
	UserAgent     string `json:"User-Agent"`
	ApVersion     string `json:"Apversion"`
	Authorization string `json:"authorization"`
}

func Get(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	userids := queries.GetIntArray(r.URL.Query(), "userids")
	eggsids := queries.GetArray(r.URL.Query(), "eggsids")
	if len(eggsids) == 0 && len(userids) == 0 {
		return logging.SE(http.StatusBadRequest, errors.New("no users specified"))
	}

	output, err := queries.GetUsers(context.Background(), eggsids, userids)
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
	t := time.Now()
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
	logging.LogFetch(req)
	resp, err := client.Do(req)
	if err != nil {
		logging.LogFetchErrored(req, resp)
		return logging.SE(http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.LogFetchErrored(req, resp)
		return logging.SE(http.StatusInternalServerError, err)
	}

	logging.LogFetchCompleted(resp, respBody, t)

	if resp.StatusCode == http.StatusUnauthorized {
		return logging.SE(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	var userRaw queries.UserRaw
	err = json.Unmarshal(respBody, &userRaw)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	user := userRaw.User()
	if !user.IsValid() {
		return logging.SE(http.StatusUnauthorized, errors.New("invalid user"))
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

	fmt.Fprint(w, `"`+token+`"`)
	return nil
}

func Delete(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	eggsid := r.URL.Query().Get("eggsid")
	if eggsid == "" {
		return logging.SE(http.StatusBadRequest, errors.New("no user specified"))
	}

	tokenAccount, se := router.AuthenticateRequestOnly(r)
	if se != nil {
		return se
	}

	if eggsid != tokenAccount {
		return logging.SE(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	err := queries.UNSAFEDeleteUser(context.Background(), eggsid)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}

	w.Write([]byte(`"Successfully deleted user ` + eggsid + `"`))
	return nil
}

func generateToken() (token string, err error) {
	token, err = queries.GenerateRandomString(64)
	logging.AddAuthenticatedUsers(1)
	return
}
