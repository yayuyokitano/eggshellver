package playlistendpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	userendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/user"
	"github.com/yayuyokitano/eggshellver/lib/queries"
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

func TestPost(t *testing.T) {
	services.Start()
	defer services.Stop()
	playlistIDs := make([]string, 0)
	//stupidly high bulk test
	for i := 0; i < 300_000; i++ {
		playlistIDs = append(playlistIDs, createPlaylistId())
	}
	b, err := json.Marshal(playlistIDs)
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

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/playlist", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?eggsIDs=%s&limit=350000", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var playlists []queries.StructuredPlaylist
	err = json.Unmarshal([]byte(w.Body.String()), &playlists)
	if err != nil {
		t.Error(err)
	}
	if len(playlists) != 300_000 {
		t.Errorf("Returned %d playlists, want %d", len(playlists), 300_000)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	var limitedPlaylists []queries.StructuredPlaylist
	err = json.Unmarshal([]byte(w.Body.String()), &limitedPlaylists)
	if err != nil {
		t.Error(err)
	}
	if len(limitedPlaylists) != 10 {
		t.Errorf("Returned %d playlists, want %d", len(limitedPlaylists), 10)
	}

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?eggsIDs=%s&limit=350000", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var emptyPlaylists []queries.StructuredPlaylist
	err = json.Unmarshal([]byte(w.Body.String()), &emptyPlaylists)
	if err != nil {
		t.Error(err)
	}
	if len(emptyPlaylists) != 0 {
		t.Errorf("Returned %d playlists, want %d", len(emptyPlaylists), 0)
	}

}

func TestGet(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/playlist", nil)
	Get(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "[]" {
		t.Errorf("Body is %s, want []", w.Body.String())
	}

	playlistIDs := make([]string, 0)
	for i := 0; i < 2; i++ {
		playlistIDs = append(playlistIDs, createPlaylistId())
	}

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(playlistIDs)
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/playlist", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}
	b, err = json.Marshal(playlistIDs[:1])
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/playlist", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token2))
	Post(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status code is %d, should be %d", w.Code, http.StatusInternalServerError)
	}

	err = services.RollbackTransaction()
	if err != nil {
		t.Fatal(err)
	}
	err = services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	token, err = userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	b, err = json.Marshal(playlistIDs)
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/playlist", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?playlistIDs=%s", playlistIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	var playlists queries.StructuredPlaylists
	err = json.Unmarshal([]byte(w.Body.String()), &playlists)
	if err != nil {
		t.Error(err)
	}

	if len(playlists) != 1 {
		t.Errorf("Returned %d playlists, want %d", len(playlists), 1)
	}
	if !playlists.Contains(queries.PartialPlaylist{EggsID: os.Getenv("TESTUSER_ID"), PlaylistID: playlistIDs[0]}) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var playlists2 queries.StructuredPlaylists
	err = json.Unmarshal([]byte(w.Body.String()), &playlists2)
	if err != nil {
		t.Error(err)
	}

	if len(playlists2) != 2 {
		t.Errorf("Returned %d playlists, want %d", len(playlists2), 2)
	}
	if !playlists2.Contains(queries.PartialPlaylist{EggsID: os.Getenv("TESTUSER_ID"), PlaylistID: playlistIDs[0]}) {
		t.Errorf("Expected slice to include %s", playlistIDs[0])
	}
	if !playlists2.Contains(queries.PartialPlaylist{EggsID: os.Getenv("TESTUSER_ID"), PlaylistID: playlistIDs[1]}) {
		t.Errorf("Expected slice to include %s", playlistIDs[1])
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?eggsIDs=%s&limit=1", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var limitedPlaylists queries.StructuredPlaylists
	err = json.Unmarshal([]byte(w.Body.String()), &limitedPlaylists)
	if err != nil {
		t.Error(err)
	}
	if len(limitedPlaylists) != 1 {
		t.Errorf("Returned %d playlists, want %d", len(limitedPlaylists), 1)
	}
	if !limitedPlaylists.Contains(queries.PartialPlaylist{EggsID: os.Getenv("TESTUSER_ID"), PlaylistID: playlistIDs[0]}) {
		t.Errorf("Expected slice to include %s", playlistIDs[0])
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/playlist?eggsIDs=%s&playlistIDs=%s", os.Getenv("TESTUSER_ID"), playlistIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var specifiedPlaylists queries.StructuredPlaylists
	err = json.Unmarshal([]byte(w.Body.String()), &specifiedPlaylists)
	if err != nil {
		t.Error(err)
	}
	if len(specifiedPlaylists) != 1 {
		t.Errorf("Returned %d playlists, want %d", len(specifiedPlaylists), 1)
	}
	if !specifiedPlaylists.Contains(queries.PartialPlaylist{EggsID: os.Getenv("TESTUSER_ID"), PlaylistID: playlistIDs[0]}) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}

}
