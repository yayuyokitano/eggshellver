package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func Get(w io.Writer, r *http.Request, _ []byte) *logging.StatusError {
	query := r.URL.Query()
	eggsID := query.Get("eggsID")
	paginator := queries.InitializePaginator(query)
	if eggsID == "" {
		return logging.SE(http.StatusBadRequest, errors.New("eggsID is required"))
	}
	timeline, err := queries.GetTimeline(context.Background(), eggsID, paginator.Offset, paginator.Limit)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	b, err := json.Marshal(timeline)
	if err != nil {
		return logging.SE(http.StatusInternalServerError, err)
	}
	w.Write(b)
	return nil
}
