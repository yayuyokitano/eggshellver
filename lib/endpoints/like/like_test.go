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

	musicIDs := make([]string, 0)
	for i := 0; i < 1; i++ {
		musicIDs = append(musicIDs, createMusicId())
	}

	b, err := json.Marshal(musicIDs)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusInternalServerError, w.Body.String())
	}

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
	var likes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	if len(likes.Likes) != 300_000 {
		t.Errorf("Returned %d liked tracks, want %d", len(likes.Likes), 300_000)
	}
	if likes.Total != 300_000 {
		t.Errorf("Returned %d total liked tracks, want %d", likes.Total, 300_000)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	var limitedLikes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &limitedLikes)
	if err != nil {
		t.Error(err)
	}
	if len(limitedLikes.Likes) != 10 {
		t.Errorf("Returned %d liked tracks, want %d", len(limitedLikes.Likes), 10)
	}
	if limitedLikes.Total != 300_000 {
		t.Errorf("Returned %d total liked tracks, want %d", limitedLikes.Total, 300_000)
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
	var emptyLikes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &emptyLikes)
	if err != nil {
		t.Error(err)
	}
	if len(emptyLikes.Likes) != 0 {
		t.Errorf("Returned %d liked tracks, want %d", len(emptyLikes.Likes), 0)
	}
	if emptyLikes.Total != 0 {
		t.Errorf("Returned %d total liked tracks, want %d", emptyLikes.Total, 0)
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
	if w.Body.String() != `{"likes":[],"total":0}` {
		t.Errorf("Body is %s, want %s", w.Body.String(), `{"likes":[],"total":0}`)
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
	if len(likes.Likes) != 2 {
		t.Errorf("Returned %d liking users, want %d", len(likes.Likes), 2)
	}
	if likes.Total != 2 {
		t.Errorf("Returned %d total likes, want %d", likes.Total, 2)
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
	if len(limitedLikes.Likes) != 1 {
		t.Errorf("Returned %d liking users, want %d", len(limitedLikes.Likes), 1)
	}
	if limitedLikes.Total != 2 {
		t.Errorf("Returned %d total likes, want %d", limitedLikes.Total, 2)
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
	if len(likes2.Likes) != 1 {
		t.Errorf("Returned %d liking users, want %d", len(likes2.Likes), 1)
	}
	if likes2.Total != 1 {
		t.Errorf("Returned %d total likes, want %d", likes2.Total, 1)
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
	if len(specifiedLikes.Likes) != 1 {
		t.Errorf("Returned %d likes, want %d", len(specifiedLikes.Likes), 1)
	}
	if specifiedLikes.Total != 1 {
		t.Errorf("Returned %d total likes, want %d", specifiedLikes.Total, 1)
	}
	if !specifiedLikes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[0]}) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}

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

	musicIDs := make([]string, 0)
	for i := 0; i < 5; i++ {
		musicIDs = append(musicIDs, createMusicId())
	}

	b, err := json.Marshal(musicIDs)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	var likes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	if likes.Total != 5 {
		t.Errorf("Returned %d total likes, want %d", likes.Total, 5)
	}

	b, err = json.Marshal(musicIDs[:1])

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/delete-like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Delete(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != fmt.Sprintf("%s unliked %d tracks", os.Getenv("TESTUSER_ID"), 1) {
		t.Errorf("Body is %s, want %s", w.Body.String(), fmt.Sprintf("%s unliked %d tracks", os.Getenv("TESTUSER_ID"), 1))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	err = json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	if likes.Total != 4 {
		t.Errorf("Returned %d total likes, want %d", likes.Total, 4)
	}
	if likes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[0]}) {
		t.Errorf("Expected slice to not include %s", musicIDs[0])
	}
	if !likes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[1]}) {
		t.Errorf("Expected slice to include %s", musicIDs[1])
	}
	if !likes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[2]}) {
		t.Errorf("Expected slice to include %s", musicIDs[2])
	}
	if !likes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[3]}) {
		t.Errorf("Expected slice to include %s", musicIDs[3])
	}
	if !likes.Contains(queries.PartialLike{EggsID: os.Getenv("TESTUSER_ID"), TrackID: musicIDs[4]}) {
		t.Errorf("Expected slice to include %s", musicIDs[4])
	}

	b, err = json.Marshal(musicIDs)

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/delete-like", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Delete(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != fmt.Sprintf("%s unliked %d tracks", os.Getenv("TESTUSER_ID"), 4) {
		t.Errorf("Body is %s, want %s", w.Body.String(), fmt.Sprintf("%s unliked %d tracks", os.Getenv("TESTUSER_ID"), 4))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}

	err = json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	if likes.Total != 0 {
		t.Errorf("Returned %d total likes, want %d", likes.Total, 0)
	}
	if len(likes.Likes) != 0 {
		t.Errorf("Returned %d likes, want %d", len(likes.Likes), 0)
	}

}
