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
	var likes queries.LikeTargetsFixed
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &likes)
	if se != nil {
		return se
	}

	if !likes.IsValid() {
		return logging.SE(http.StatusBadRequest, errors.New("invalid likedTracks"))
	}

	n, err := queries.LikeObjects(context.Background(), eggsID, likes)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	logging.AddLikes(int(n), likes.Type)
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

func Put(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var likes queries.LikeTargetsFixed
	eggsID, se := router.AuthenticatePostRequest(w, r, b, &likes)
	if se != nil {
		return se
	}

	if !likes.IsValid() {
		return logging.SE(http.StatusBadRequest, errors.New("invalid likes"))
	}

	delta, total, err := queries.PutLikes(context.Background(), eggsID, likes)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	logging.AddLikes(int(delta), likes.Type)
	fmt.Fprint(w, total)
	return nil
}

func Toggle(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var targetID string
	var targetType string
	eggsID, se := router.AuthenticateIndividualPostRequest(w, r, b, &targetType, &targetID)
	if se != nil {
		return se
	}

	target := queries.LikeTarget{
		ID:   targetID,
		Type: targetType,
	}
	if !target.IsValid() {
		return logging.SE(http.StatusBadRequest, errors.New("invalid target"))
	}

	isLiking, err := queries.ToggleLike(context.Background(), eggsID, target)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	if isLiking {
		logging.AddLikes(1, target.Type)
	} else {
		logging.AddLikes(-1, target.Type)
	}
	fmt.Fprint(w, isLiking)
	return nil
}
