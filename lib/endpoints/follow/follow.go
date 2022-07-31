package followendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

func Delete(w http.ResponseWriter, r *http.Request) {
	var unfollowedUsers []string
	eggsID, err := router.AuthenticatePostRequest(w, r, &unfollowedUsers)
	if err != nil {
		return
	}

	n, err := queries.DeleteFollows(context.Background(), eggsID, unfollowedUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%s unfollowed %d users", eggsID, n)))
}

func Post(w http.ResponseWriter, r *http.Request) {
	var followedUsers []string
	eggsID, err := router.AuthenticatePostRequest(w, r, &followedUsers)
	if err != nil {
		return
	}

	n, err := queries.SubmitFollows(context.Background(), eggsID, followedUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%s followed %d users", eggsID, n)))
}

func Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	followerIDs := queries.GetArray(query, "followerIDs")
	followeeIDs := queries.GetArray(query, "followeeIDs")
	paginator := queries.InitializePaginator(query)
	if len(followerIDs) == 0 && len(followeeIDs) == 0 {
		http.Error(w, "followerIDs and/or followeeIDs is required", http.StatusBadRequest)
	}
	follows, err := queries.GetFollows(context.Background(), followerIDs, followeeIDs, paginator)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(follows)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
