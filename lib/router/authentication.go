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

func AuthenticateSpecificUser(r *http.Request) (userStub queries.UserStub, se *logging.StatusError) {
	authSplit := strings.Split(r.URL.Path, "/")
	user := authSplit[len(authSplit)-2]

	if user == "" {
		se = logging.SE(http.StatusBadRequest, errors.New("please specify user to authenticate"))
		return
	}

	userStub, se = UserStubFromToken(r)
	if se != nil {
		return
	}
	if userStub.EggsID != user {
		se = logging.SE(http.StatusUnauthorized, errors.New("failed to authenticate user"))
	}

	return
}

func UserStubFromToken(r *http.Request) (userStub queries.UserStub, statusErr *logging.StatusError) {
	authSplit := strings.Split(r.URL.Path, "/")
	token := authSplit[len(authSplit)-1]

	userStubs, err := queries.GetUserStubFromToken(context.Background(), token)
	if err != nil {
		statusErr = logging.SE(http.StatusUnauthorized, err)
		return
	}
	if len(userStubs) == 0 {
		statusErr = logging.SE(http.StatusUnauthorized, errors.New("failed to authenticate user"))
		return
	}
	userStub = userStubs[0]
	return
}

func authenticateUser(bearer string) (eggsID string, err error) {
	eggsID, err = queries.GetEggsIDByToken(context.Background(), bearer[7:])
	return
}
