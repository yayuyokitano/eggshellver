package userstubendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var users []queries.UserStub
	err := json.NewDecoder(r.Body).Decode(&users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	n, err := queries.InsertUserStubs(context.Background(), users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("%d users inserted", n)))
}
