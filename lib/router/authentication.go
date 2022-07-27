package router

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func AuthenticatePostRequest[V any](w http.ResponseWriter, r *http.Request, v *V) (eggsID string, err error) {
	err = json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	eggsID, err = authenticateUser(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	return
}

func authenticateUser(bearer string) (eggsID string, err error) {
	eggsID, err = queries.GetEggsIDByToken(context.Background(), bearer[7:])
	return
}
