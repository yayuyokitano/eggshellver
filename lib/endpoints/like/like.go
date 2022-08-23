package likeendpoint

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
	var likedTracks queries.LikeTargets
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &likedTracks)
	if se != nil {
		return se
	}

	if !likedTracks.IsValid() {
		return logging.SE(http.StatusBadRequest, errors.New("invalid likedTracks"))
	}

	n, err := queries.LikeObjects(context.Background(), eggsID, likedTracks)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}

func Get(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	query := r.URL.Query()
	eggsIDs := queries.GetArray(query, "eggsIDs")
	targetIDs := queries.GetArray(query, "targetIDs")
	targetType := r.URL.Query().Get("targetType")
	paginator := queries.InitializePaginator(query)
	if len(eggsIDs) == 0 && len(targetIDs) == 0 {
		return logging.SE(http.StatusBadRequest, errors.New("eggsIDs and/or targetIDs is required"))
	}
	likedTracks, err := queries.GetLikedObjects(context.Background(), eggsIDs, targetIDs, targetType, paginator)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	b, err := json.Marshal(likedTracks)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	w.Write(b)
	return nil
}

func Delete(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	var deletedLikes []string
	eggsID, se := router.AuthenticateDeleteRequest(w, r, &deletedLikes)
	if se != nil {
		return se
	}

	n, err := queries.DeleteLikes(context.Background(), eggsID, deletedLikes)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}

func Put(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var likes queries.LikeTargetsFixed
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &likes)
	if se != nil {
		return se
	}

	if !likes.IsValid() {
		return logging.SE(http.StatusBadRequest, errors.New("invalid likes"))
	}

	n, err := queries.PutLikes(context.Background(), eggsID, likes)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}
