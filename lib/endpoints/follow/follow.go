package followendpoint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/router"
)

type TrackStub struct {
	MusicID string `json:"musicId"`
}

type LikedTracks struct {
	Tracks []queries.Like `json:"data"`
	Count  int            `json:"totalCount"`
}

func Post(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var followedUsers []string
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &followedUsers)
	if se != nil {
		return se
	}

	n, err := queries.SubmitFollows(context.Background(), eggsID, followedUsers)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}

func Get(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	query := r.URL.Query()
	followerIDs := queries.GetArray(query, "followerIDs")
	followeeIDs := queries.GetArray(query, "followeeIDs")
	paginator := queries.InitializePaginator(query)
	if len(followerIDs) == 0 && len(followeeIDs) == 0 {
		return logging.SE(http.StatusBadRequest, errors.New("followerIDs and/or followeeIDs is required"))
	}
	follows, err := queries.GetFollows(context.Background(), followerIDs, followeeIDs, paginator)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	b, err := json.Marshal(follows)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	w.Write(b)
	return nil
}

func Put(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var follows []string
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &follows)
	if se != nil {
		return se
	}

	n, err := queries.PutFollows(context.Background(), eggsID, follows)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}

func ToggleFollow(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var follow string
	eggsID, se := router.AuthenticateIndividualPostRequest(w, r, b, &follow)
	if se != nil {
		return se
	}

	isFollowing, err := queries.ToggleFollow(context.Background(), eggsID, follow)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, isFollowing)
	return nil
}
