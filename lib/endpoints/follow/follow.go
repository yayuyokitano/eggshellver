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
	eggsID, se := router.AuthenticatePostRequest(r, b, &followedUsers)
	if se != nil {
		return se
	}
	if eggsID == "" {
		return logging.SE(http.StatusBadRequest, errors.New("eggsID is required"))
	}
	if len(followedUsers) == 0 {
		fmt.Fprint(w, 0)
		return nil
	}

	n, err := queries.SubmitFollows(context.Background(), eggsID, followedUsers)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	logging.AddFollows(int(n))
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
	eggsID, se := router.AuthenticatePostRequest(r, b, &follows)
	if se != nil {
		return se
	}
	if eggsID == "" {
		return logging.SE(http.StatusBadRequest, errors.New("eggsID is required"))
	}
	if len(follows) == 0 {
		fmt.Fprint(w, 0)
		return nil
	}

	delta, total, err := queries.PutFollows(context.Background(), eggsID, follows)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	logging.AddFollows(int(delta))
	fmt.Fprint(w, total)
	return nil
}

func Toggle(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var follow string
	eggsID, se := router.AuthenticateIndividualPostRequest(r, b, &follow)
	if se != nil {
		return se
	}
	if eggsID == "" {
		return logging.SE(http.StatusBadRequest, errors.New("eggsID is required"))
	}
	if follow == "" {
		return logging.SE(http.StatusBadRequest, errors.New("follow is required"))
	}

	isFollowing, err := queries.ToggleFollow(context.Background(), eggsID, follow)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	if isFollowing {
		logging.AddFollows(1)
	} else {
		logging.AddFollows(-1)
	}
	fmt.Fprint(w, isFollowing)
	return nil
}
