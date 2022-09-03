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

	likes := queries.LikeTargets{{
		ID:   createMusicId(),
		Type: "track",
	}}

	b, err := json.Marshal(likes)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 1)

	likes[0].Type = "playlist"

	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 0)

}

func TestPost(t *testing.T) {
	services.Start()
	defer services.Stop()
	likes := make(queries.LikeTargets, 0)
	bulkSize := 10_000
	for i := 0; i < bulkSize; i++ {
		if i%5 == 0 {
			likes = append(likes, queries.LikeTarget{
				ID:   createMusicId(),
				Type: "track",
			})
		} else {
			likes = append(likes, queries.LikeTarget{
				ID:   createMusicId(),
				Type: "playlist",
			})
		}
	}
	b, err := json.Marshal(likes)
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

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetType=track&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, bulkSize/5, int64(bulkSize/5), []queries.PartialLike{})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetType=playlist&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 4*bulkSize/5, int64(4*bulkSize/5), []queries.PartialLike{})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 10, int64(bulkSize), []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: likes[1].ID,
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
	router.HandleMethod(Get, w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	router.HandleMethod(Get, w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != `{"likes":[],"total":0}` {
		t.Errorf("Body is %s, want %s", w.Body.String(), `{"likes":[],"total":0}`)
	}

	likeTargets := queries.LikeTargets{
		{
			ID:   createMusicId(),
			Type: "track",
		},
		{
			ID:   createMusicId(),
			Type: "playlist",
		},
	}

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(likeTargets)
	if err != nil {
		t.Error(err)
	}
	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 2)

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}
	b, err = json.Marshal(likeTargets[:1])
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token2, 1)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?targetIDs=%s", likeTargets[0].ID), nil)
	testHasLikes(t, r, 2, 2, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: likeTargets[0].ID,
	}, {
		EggsID:   os.Getenv("TESTUSER_ID2"),
		TargetID: likeTargets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?targetIDs=%s&limit=1", likeTargets[0].ID), nil)
	testHasLikes(t, r, 1, 2, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID2"),
		TargetID: likeTargets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?targetIDs=%s", likeTargets[1].ID), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: likeTargets[1].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetIDs=%s", os.Getenv("TESTUSER_ID"), likeTargets[0].ID), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: likeTargets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetType=%s", os.Getenv("TESTUSER_ID"), "track"), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: likeTargets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetType=%s", os.Getenv("TESTUSER_ID2"), "playlist"), nil)
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

	likeTargets := make(queries.LikeTargets, 0)
	for i := 0; i < 5; i++ {
		likeTargets = append(likeTargets, queries.LikeTarget{
			ID:   createMusicId(),
			Type: "track",
		})
	}

	playlistTarget := queries.LikeTarget{
		ID:   createMusicId(),
		Type: "playlist",
	}

	b, err := json.Marshal(queries.LikeTargetsFixed{
		Targets: likeTargets[:3],
		Type:    "track",
	})
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("PUT", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Put, token, 3)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikedIDs(t, r, 3, 3, os.Getenv("TESTUSER_ID"), likeTargets.IDs()[:3])

	b, err = json.Marshal(queries.LikeTargetsFixed{
		Targets: likeTargets[1:],
		Type:    "track",
	})
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("PUT", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Put, token, 4)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikedIDs(t, r, 4, 4, os.Getenv("TESTUSER_ID"), likeTargets.IDs()[1:])

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	router.HandleMethod(Get, w, r)
	var likes queries.StructuredLikes
	err = json.Unmarshal(w.Body.Bytes(), &likes)
	if err != nil {
		t.Error(err)
	}
	//Timestamps should not have changed for the tracks that had been recorded
	if likes.Likes[0].ID != likeTargets[3].ID {
		t.Errorf("Expected music ID to be %s, got %s", likeTargets[3].ID, likes.Likes[0].ID)
	}
	if likes.Likes[2].ID != likeTargets[1].ID {
		t.Errorf("Expected music ID to be %s, got %s", likeTargets[1].ID, likes.Likes[2].ID)
	}

	b, err = json.Marshal(queries.LikeTargetsFixed{
		Targets: []queries.LikeTarget{playlistTarget},
		Type:    "playlist",
	})
	if err != nil {
		t.Error(err)
	}
	r = httptest.NewRequest("PUT", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Put, token, 1)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikedIDs(t, r, 5, 5, os.Getenv("TESTUSER_ID"), append(likeTargets.IDs()[1:], playlistTarget.ID))
}

func TestToggle(t *testing.T) {
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

	likeTarget := createMusicId()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/like", nil)
	router.HandleMethod(Toggle, w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code to be %d, got %d", http.StatusBadRequest, w.Code)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/like/", nil)
	router.HandleMethod(Toggle, w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code to be %d, got %d", http.StatusBadRequest, w.Code)
	}

	r = httptest.NewRequest("POST", fmt.Sprintf("/like/track/%s", likeTarget), nil)
	router.CommitToggling(t, r, Toggle, token, true)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		TargetID: likeTarget,
		EggsID:   os.Getenv("TESTUSER_ID"),
	}})

	r = httptest.NewRequest("POST", fmt.Sprintf("/like/track/%s", likeTarget), nil)
	router.CommitToggling(t, r, Toggle, token, false)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 0, 0, []queries.PartialLike{})

}

func testHasLikedIDs(t *testing.T, r *http.Request, num int, total int64, eggsID string, expectedLikedIDs []string) {
	expectedLikes := make([]queries.PartialLike, 0)
	for _, trackID := range expectedLikedIDs {
		expectedLikes = append(expectedLikes, queries.PartialLike{
			EggsID:   eggsID,
			TargetID: trackID,
		})
	}
	testHasLikes(t, r, num, total, expectedLikes)
}

func testHasLikes(t *testing.T, r *http.Request, num int, total int64, expectedLikes []queries.PartialLike) {
	t.Helper()
	w := httptest.NewRecorder()
	router.HandleMethod(Get, w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}

	var likes queries.StructuredLikes
	err := json.Unmarshal(w.Body.Bytes(), &likes)
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
