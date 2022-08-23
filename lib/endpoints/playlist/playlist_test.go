package playlistendpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	userendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/user"
	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/router"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

func generateStringPanic(n int) string {
	b, err := queries.GenerateRandomString(n)
	if err != nil {
		panic(err)
	}
	return b
}

func createPlaylistId() string {
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		generateStringPanic(8),
		generateStringPanic(4),
		generateStringPanic(4),
		generateStringPanic(4),
		generateStringPanic(12),
	)
}

func TestUniquePost(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	playlists := make([]queries.PlaylistInput, 0)
	for i := 0; i < 1; i++ {
		playlists = append(playlists, queries.PlaylistInput{
			PlaylistID:   createPlaylistId(),
			LastModified: time.Now().Add(time.Duration(-i) * time.Second),
		})
	}

	b, err := json.Marshal(playlists)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("POST", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 1)

	r = httptest.NewRequest("POST", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 1)

}

func TestPost(t *testing.T) {
	services.Start()
	defer services.Stop()
	playlists := make(queries.PlaylistInputs, 0)
	bulkSize := 10_000
	for i := 0; i < bulkSize; i++ {
		playlists = append(playlists, queries.PlaylistInput{
			PlaylistID:   createPlaylistId(),
			LastModified: time.Now().Add(time.Duration(-i) * time.Second),
		})
	}
	b, err := json.Marshal(playlists)
	if err != nil {
		t.Error(err)
	}

	err = services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("POST", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, int64(bulkSize))

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, bulkSize, int64(bulkSize), []queries.PartialPlaylist{})

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 10, int64(bulkSize), []queries.PartialPlaylist{{
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[0].PlaylistID,
	}})

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 0, 0, []queries.PartialPlaylist{})

}

func TestGet(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/playlists", nil)
	router.HandleMethod(Get, w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	router.HandleMethod(Get, w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != `{"playlists":[],"total":0}` {
		t.Errorf("Body is %s, want %s", w.Body.String(), `{"playlists":[],"total":0}`)
	}

	playlists := make([]queries.PlaylistInput, 0)
	for i := 0; i < 2; i++ {
		playlists = append(playlists, queries.PlaylistInput{
			PlaylistID:   createPlaylistId(),
			LastModified: time.Now().Add(time.Duration(-i) * time.Second),
		})
	}

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(playlists)
	if err != nil {
		t.Error(err)
	}
	r = httptest.NewRequest("POST", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, int64(len(playlists)))

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}
	b, err = json.Marshal(playlists[:1])
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("POST", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token2, 1)

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?playlistIDs=%s", playlists[0].PlaylistID), nil)
	testHasPlaylists(t, r, 1, 1, []queries.PartialPlaylist{{
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[0].PlaylistID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 2, 2, []queries.PartialPlaylist{{
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[0].PlaylistID,
	}, {
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[1].PlaylistID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s&limit=1", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 1, 2, []queries.PartialPlaylist{{
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[0].PlaylistID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s&playlistIDs=%s", os.Getenv("TESTUSER_ID"), playlists[1].PlaylistID), nil)
	testHasPlaylists(t, r, 1, 1, []queries.PartialPlaylist{{
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[1].PlaylistID,
	}})
}

func TestDelete(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	playlists := make(queries.PlaylistInputs, 0)
	for i := 0; i < 5; i++ {
		playlists = append(playlists, queries.PlaylistInput{
			PlaylistID:   createPlaylistId(),
			LastModified: time.Now().Add(time.Duration(-i) * time.Second),
		})
	}

	b, err := json.Marshal(playlists)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("POST", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, int64(len(playlists)))

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 5, 5, []queries.PartialPlaylist{})

	r = httptest.NewRequest("DELETE", fmt.Sprintf("/playlists?target=%s", playlists[0].PlaylistID), nil)
	router.CommitMutating(t, r, Delete, token, 1)

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 4, 4, []queries.PartialPlaylist{{
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[1].PlaylistID,
	}, {
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[2].PlaylistID,
	}, {
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[3].PlaylistID,
	}, {
		EggsID:     os.Getenv("TESTUSER_ID"),
		PlaylistID: playlists[4].PlaylistID,
	}})

	s := strings.Join(playlists.PlaylistIDs(), ",")

	r = httptest.NewRequest("DELETE", fmt.Sprintf("/playlists?target=%s", s), nil)
	router.CommitMutating(t, r, Delete, token, 4)

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 0, 0, []queries.PartialPlaylist{})

}

func TestPut(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	playlists := make(queries.PlaylistInputs, 0)
	for i := 0; i < 5; i++ {
		playlists = append(playlists, queries.PlaylistInput{
			PlaylistID:   createPlaylistId(),
			LastModified: time.Now().Add(time.Duration(-i) * time.Second),
		})
	}

	b, err := json.Marshal(playlists[:3])
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("PUT", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Put, token, 3)

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 3, 3, playlists.PartialPlaylists(os.Getenv("TESTUSER_ID"))[:3])

	b, err = json.Marshal(playlists[1:])
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("PUT", "/playlists", bytes.NewReader(b))
	router.CommitMutating(t, r, Put, token, 4)

	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasPlaylists(t, r, 4, 4, playlists.PartialPlaylists(os.Getenv("TESTUSER_ID"))[1:])

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlists?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	router.HandleMethod(Get, w, r)

	var playlistResults queries.StructuredPlaylists
	err = json.Unmarshal(w.Body.Bytes(), &playlistResults)
	if err != nil {
		t.Error(err)
	}
	//Timestamps SHOULD have changed for the playlists that had been recorded
	if playlistResults.Playlists[0].PlaylistID != playlists[1].PlaylistID {
		t.Errorf("Expected music ID to be %s, got %s", playlists[1].PlaylistID, playlistResults.Playlists[0].PlaylistID)
	}
	if playlistResults.Playlists[2].PlaylistID != playlists[3].PlaylistID {
		t.Errorf("Expected music ID to be %s, got %s", playlists[3].PlaylistID, playlistResults.Playlists[2].PlaylistID)
	}
}

func testHasPlaylists(t *testing.T, r *http.Request, num int, total int64, expectedPlaylists []queries.PartialPlaylist) {
	t.Helper()
	w := httptest.NewRecorder()
	router.HandleMethod(Get, w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}

	var playlists queries.StructuredPlaylists
	err := json.Unmarshal(w.Body.Bytes(), &playlists)
	if err != nil {
		t.Error(err)
	}
	if len(playlists.Playlists) != num {
		t.Errorf("Returned %d playlists, want %d", len(playlists.Playlists), num)
	}
	if playlists.Total != total {
		t.Errorf("Returned %d total, want %d", playlists.Total, total)
	}
	for _, playlist := range expectedPlaylists {
		if !playlists.Contains(playlist) {
			t.Errorf("Expected slice to include playlist %v", playlist)
		}
	}
}
