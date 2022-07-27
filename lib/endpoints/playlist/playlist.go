package playlistendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/router"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var playlists []string
	eggsID, err := router.AuthenticatePostRequest(w, r, &playlists)
	if err != nil {
		return
	}

	n, err := queries.PostPlaylists(context.Background(), eggsID, playlists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%s attached %d playlists", eggsID, n)))
}

func Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	eggsIDs := queries.GetArray(query, "eggsIDs")
	playlistIDs := queries.GetArray(query, "playlistIDs")
	paginator := queries.InitializePaginator(query)
	if len(eggsIDs) == 0 && len(playlistIDs) == 0 {
		http.Error(w, "eggsIDs and/or playlistIDs is required", http.StatusBadRequest)
	}
	playlists, err := queries.GetPlaylists(context.Background(), eggsIDs, playlistIDs, paginator)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(playlists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
