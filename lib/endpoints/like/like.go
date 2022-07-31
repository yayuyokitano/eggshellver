package likeendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/router"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	var deletedLikes []string
	eggsID, err := router.AuthenticatePostRequest(w, r, &deletedLikes)
	if err != nil {
		return
	}

	n, err := queries.DeleteLikes(context.Background(), eggsID, deletedLikes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%s unliked %d tracks", eggsID, n)))
}

func Post(w http.ResponseWriter, r *http.Request) {
	var likedTracks []string
	eggsID, err := router.AuthenticatePostRequest(w, r, &likedTracks)
	if err != nil {
		return
	}

	n, err := queries.LikeTracks(context.Background(), eggsID, likedTracks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%s liked %d tracks", eggsID, n)))
}

func Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	eggsIDs := queries.GetArray(query, "eggsIDs")
	trackIDs := queries.GetArray(query, "trackIDs")
	paginator := queries.InitializePaginator(query)
	if len(eggsIDs) == 0 && len(trackIDs) == 0 {
		http.Error(w, "eggsIDs and/or trackIDs is required", http.StatusBadRequest)
	}
	likedTracks, err := queries.GetLikedTracks(context.Background(), eggsIDs, trackIDs, paginator)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(likedTracks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
