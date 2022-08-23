package router

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func AuthenticatePostRequest[V any](w io.Writer, r *http.Request, b []byte, v *V) (eggsID string, statusErr *logging.StatusError) {
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

func AuthenticateDeleteRequest(w io.Writer, r *http.Request, v *[]string) (eggsID string, statusErr *logging.StatusError) {
	*v = queries.GetArray(r.URL.Query(), "target")
	eggsID, err := authenticateUser(r.Header.Get("Authorization"))
	if err != nil {
		statusErr = logging.SE(http.StatusUnauthorized, err)
	}
	return
}

func authenticateUser(bearer string) (eggsID string, err error) {
	eggsID, err = queries.GetEggsIDByToken(context.Background(), bearer[7:])
	return
}
