package likeendpoint

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

	r := httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 1)

	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 0)

}

func TestPost(t *testing.T) {
	services.Start()
	defer services.Stop()
	musicIDs := make([]string, 0)
	bulkSize := 10_000
	for i := 0; i < bulkSize; i++ {
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

	r := httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, int64(bulkSize))

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, bulkSize, int64(bulkSize), []queries.PartialLike{})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 10, int64(bulkSize), []queries.PartialLike{{
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[1],
	}})

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 0, 0, []queries.PartialLike{})

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
	r := httptest.NewRequest("GET", "/likes", nil)
	Get(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
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
	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 2)

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}
	b, err = json.Marshal(musicIDs[:1])
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token2, 1)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?trackIDs=%s", musicIDs[0]), nil)
	testHasLikes(t, r, 2, 2, []queries.PartialLike{{
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[0],
	}, {
		EggsID:  os.Getenv("TESTUSER_ID2"),
		TrackID: musicIDs[0],
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?trackIDs=%s&limit=1", musicIDs[0]), nil)
	testHasLikes(t, r, 1, 2, []queries.PartialLike{{
		EggsID:  os.Getenv("TESTUSER_ID2"),
		TrackID: musicIDs[0],
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?trackIDs=%s", musicIDs[1]), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[1],
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&trackIDs=%s", os.Getenv("TESTUSER_ID"), musicIDs[0]), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[0],
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

	musicIDs := make([]string, 0)
	for i := 0; i < 5; i++ {
		musicIDs = append(musicIDs, createMusicId())
	}

	b, err := json.Marshal(musicIDs)

	r := httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, int64(len(musicIDs)))

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 5, 5, []queries.PartialLike{})

	r = httptest.NewRequest("DELETE", fmt.Sprintf("/likes?target=%s", musicIDs[0]), nil)
	router.CommitMutating(t, r, Delete, token, 1)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 4, 4, []queries.PartialLike{{
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[1],
	}, {
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[2],
	}, {
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[3],
	}, {
		EggsID:  os.Getenv("TESTUSER_ID"),
		TrackID: musicIDs[4],
	}})

	s := strings.Join(musicIDs, ",")

	r = httptest.NewRequest("DELETE", fmt.Sprintf("/likes?target=%s", s), nil)
	router.CommitMutating(t, r, Delete, token, 4)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 0, 0, []queries.PartialLike{})

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

	musicIDs := make([]string, 0)
	for i := 0; i < 5; i++ {
		musicIDs = append(musicIDs, createMusicId())
	}

	b, err := json.Marshal(musicIDs[:3])

	r := httptest.NewRequest("PUT", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Put, token, 3)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikedIDs(t, r, 3, 3, os.Getenv("TESTUSER_ID"), musicIDs[:3])

	b, err = json.Marshal(musicIDs[1:])

	r = httptest.NewRequest("PUT", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Put, token, 4)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikedIDs(t, r, 4, 4, os.Getenv("TESTUSER_ID"), musicIDs[1:])

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)
	var likes queries.StructuredLikes
	err = json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	//Timestamps should not have changed for the tracks that had been recorded
	if likes.Likes[0].TrackID != musicIDs[3] {
		t.Errorf("Expected music ID to be %s, got %s", musicIDs[3], likes.Likes[0].TrackID)
	}
	if likes.Likes[2].TrackID != musicIDs[1] {
		t.Errorf("Expected music ID to be %s, got %s", musicIDs[1], likes.Likes[2].TrackID)
	}
}

func testHasLikedIDs(t *testing.T, r *http.Request, num int, total int64, eggsID string, expectedLikedIDs []string) {
	expectedLikes := make([]queries.PartialLike, 0)
	for _, trackID := range expectedLikedIDs {
		expectedLikes = append(expectedLikes, queries.PartialLike{
			EggsID:  eggsID,
			TrackID: trackID,
		})
	}
	testHasLikes(t, r, num, total, expectedLikes)
}

func testHasLikes(t *testing.T, r *http.Request, num int, total int64, expectedLikes []queries.PartialLike) {
	t.Helper()
	w := httptest.NewRecorder()
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}

	var likes queries.StructuredLikes
	err := json.Unmarshal([]byte(w.Body.String()), &likes)
	if err != nil {
		t.Error(err)
	}
	if len(likes.Likes) != num {
		t.Errorf("Returned %d liked tracks, want %d", len(likes.Likes), num)
	}
	if likes.Total != total {
		t.Errorf("Returned %d total, want %d", likes.Total, total)
	}
	for _, like := range expectedLikes {
		if !likes.Contains(like) {
			t.Errorf("Expected slice to include like %v", like)
		}
	}
}
