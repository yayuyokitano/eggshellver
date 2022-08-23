package userstubendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func Post(w io.Writer, r *http.Request, b []byte) *logging.StatusError {
	var users []queries.UserStub
	err := json.Unmarshal(b, &users)
	if err != nil {
		return logging.SE(http.StatusBadRequest, err)
	}
	n, err := queries.PostUserStubs(context.Background(), users)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	fmt.Fprint(w, n)
	return nil
}
