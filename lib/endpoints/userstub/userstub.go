package userstubendpoint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func Post(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var users queries.UserStubs
	err := json.Unmarshal(b, &users)
	if err != nil {
		return logging.SE(http.StatusBadRequest, err)
	}
	if !users.IsValid() {
		return logging.SE(http.StatusBadRequest, errors.New("invalid user stubs"))
	}
	inserted, updated, err := queries.PostUserStubs(context.Background(), users)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	logging.AddCachedUsers(int(inserted))
	fmt.Fprint(w, inserted+updated)
	return nil
}
