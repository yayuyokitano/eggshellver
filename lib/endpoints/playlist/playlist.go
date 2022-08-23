package playlistendpoint

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

func Post(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var playlists []queries.PlaylistInput
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &playlists)
	if se != nil {
		return se
	}

	n, err := queries.PostPlaylists(context.Background(), eggsID, playlists)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}

func Get(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	query := r.URL.Query()
	eggsIDs := queries.GetArray(query, "eggsIDs")
	playlistIDs := queries.GetArray(query, "playlistIDs")
	paginator := queries.InitializePaginator(query)
	if len(eggsIDs) == 0 && len(playlistIDs) == 0 {
		return logging.SE(http.StatusBadRequest, errors.New("eggsIDs and/or playlistIDs is required"))
	}
	playlists, err := queries.GetPlaylists(context.Background(), eggsIDs, playlistIDs, paginator)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	b, err := json.Marshal(playlists)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	w.Write(b)
	return nil
}

func Delete(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	var deletedPlaylists []string
	eggsID, se := router.AuthenticateDeleteRequest(w, r, &deletedPlaylists)
	if se != nil {
		return se
	}

	n, err := queries.DeletePlaylists(context.Background(), eggsID, deletedPlaylists)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}

func Put(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var playlists queries.PlaylistInputs
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &playlists)
	if se != nil {
		return se
	}

	n, err := queries.PutPlaylists(context.Background(), eggsID, playlists)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}
