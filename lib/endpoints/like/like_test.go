package likeendpoint

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

func createMusicId() string {
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
	musicIDs := make([]string, 0)
	//stupidly high bulk test
	for i := 0; i < 300_000; i++ {
		musicIDs = append(musicIDs, createMusicId())
	}
	b, err := json.Marshal(musicIDs)
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
	r := httptest.NewRequest("POST", "/like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s&limit=350000", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var likes []queries.StructuredLike
	err = json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	if len(likes) != 300_000 {
		t.Errorf("Returned %d liked tracks, want %d", len(likes), 300_000)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	var limitedLikes []queries.StructuredLike
	err = json.Unmarshal([]byte(w.Body.String()), &limitedLikes)
	if err != nil {
		t.Error(err)
	}
	if len(limitedLikes) != 10 {
		t.Errorf("Returned %d liked tracks, want %d", len(limitedLikes), 10)
	}

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s&limit=350000", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var emptyLikes []queries.StructuredLike
	err = json.Unmarshal([]byte(w.Body.String()), &emptyLikes)
	if err != nil {
		t.Error(err)
	}
	if len(emptyLikes) != 0 {
		t.Errorf("Returned %d liked tracks, want %d", len(emptyLikes), 0)
	}

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
	r := httptest.NewRequest("GET", "/like", nil)
	Get(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "[]" {
		t.Errorf("Body is %s, want []", w.Body.String())
	}

	musicIDs := make([]string, 0)
	for i := 0; i < 2; i++ {
		musicIDs = append(musicIDs, createMusicId())
	}

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(musicIDs)
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}
	b, err = json.Marshal(musicIDs[:1])
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token2))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?trackIDs=%s", musicIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	var likes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	if len(likes) != 2 {
		t.Errorf("Returned %d liking users, want %d", len(likes), 2)
	}
	if !likes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[0]}) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}
	if !likes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID2"), TrackID: musicIDs[0]}) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID2"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?trackIDs=%s&limit=1", musicIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var limitedLikes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &limitedLikes)
	if err != nil {
		t.Error(err)
	}
	if len(limitedLikes) != 1 {
		t.Errorf("Returned %d liking users, want %d", len(limitedLikes), 1)
	}
	if !limitedLikes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[0]}) && !limitedLikes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID2"), TrackID: musicIDs[0]}) {
		t.Errorf("Expected slice to include %s or %s", os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?trackIDs=%s", musicIDs[1]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var likes2 queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &likes2)
	if err != nil {
		t.Error(err)
	}
	if len(likes2) != 1 {
		t.Errorf("Returned %d liking users, want %d", len(likes2), 1)
	}
	if !likes2.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[1]}) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s&trackIDs=%s", os.Getenv("TESTUSER_ID"), musicIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var specifiedLikes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &specifiedLikes)
	if err != nil {
		t.Error(err)
	}
	if len(specifiedLikes) != 1 {
		t.Errorf("Returned %d likes, want %d", len(specifiedLikes), 1)
	}
	if !specifiedLikes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[0]}) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}

}
