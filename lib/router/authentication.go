package router

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func AuthenticatePostRequest[V any](r *http.Request, b []byte, v *V) (eggsID string, statusErr *logging.StatusError) {
	err := json.Unmarshal(b, v)
	if err != nil {
		statusErr = logging.SE(http.StatusBadRequest, err)
		return
	}
	eggsID, err = authenticateUser(r.Header.Get("Authorization"))
	if err != nil {
		statusErr = logging.SE(http.StatusUnauthorized, err)
	}
	return
}

func AuthenticateIndividualPostRequest(r *http.Request, b []byte, v ...*string) (eggsID string, statusErr *logging.StatusError) {
	pathSplit := strings.Split(r.URL.Path, "/")
	if len(pathSplit) < len(v)+2 {
		statusErr = logging.SE(http.StatusBadRequest, errors.New("invalid path"))
		return
	}
	for i := 0; i < len(v); i++ {
		*v[i] = pathSplit[i+2]
		if *v[i] == "" {
			statusErr = logging.SE(http.StatusBadRequest, errors.New("invalid path"))
			return
		}
	}
	eggsID, err := authenticateUser(r.Header.Get("Authorization"))
	if err != nil {
		statusErr = logging.SE(http.StatusUnauthorized, err)
	}
	return
}

func AuthenticateDeleteRequest(r *http.Request, v *[]string) (eggsID string, statusErr *logging.StatusError) {
	*v = queries.GetArray(r.URL.Query(), "target")
	eggsID, err := authenticateUser(r.Header.Get("Authorization"))
	if err != nil {
		statusErr = logging.SE(http.StatusUnauthorized, err)
	}
	return
}

func AuthenticateSpecificUser(r *http.Request, user string) (statusErr *logging.StatusError) {
	eggsID, err := authenticateUser(r.Header.Get("Authorization"))
	if err != nil || eggsID != user {
		statusErr = logging.SE(http.StatusUnauthorized, err)
	}
	return
}

func authenticateUser(bearer string) (eggsID string, err error) {
	eggsID, err = queries.GetEggsIDByToken(context.Background(), bearer[7:])
	return
}
