package userendpoint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/logging"
)

type TwitterOauth struct {
	OauthToken    string `json:"oauth_token"`
	OauthVerifier string `json:"oauth_verifier"`
}

func twitterOauthUrl(twitterOauth TwitterOauth) string {
	return fmt.Sprintf("https://api.twitter.com/oauth/access_token?oauth_token=%s&oauth_verifier=%s", twitterOauth.OauthToken, twitterOauth.OauthVerifier)
}

func PostTwitter(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var twitterOauth TwitterOauth
	err := json.Unmarshal(b, &twitterOauth)
	if err != nil {
		return logging.SE(http.StatusBadRequest, err)
	}

	resp, err := http.Post(twitterOauthUrl(twitterOauth), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	w.Write(b)
	return nil
}
