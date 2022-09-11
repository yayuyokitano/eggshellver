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

	likes := queries.LikeTargetsFixed{
		Targets: []queries.LikeTarget{{
			ID:   createMusicId(),
			Type: "track",
		}},
		Type: "track",
	}

	b, err := json.Marshal(likes)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 1)

	likes.Targets[0].Type = "playlist"
	likes.Type = "playlist"

	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 0)

}

func TestPost(t *testing.T) {
	services.Start()
	defer services.Stop()
	trackLikes := queries.LikeTargetsFixed{
		Targets: make(queries.LikeTargets, 0),
		Type:    "track",
	}
	playlistLikes := queries.LikeTargetsFixed{
		Targets: make(queries.LikeTargets, 0),
		Type:    "playlist",
	}
	bulkSize := 10_000
	for i := 0; i < bulkSize; i++ {
		if i%5 == 0 {
			trackLikes.Targets = append(trackLikes.Targets, queries.LikeTarget{
				ID:   createMusicId(),
				Type: "track",
			})
		} else {
			playlistLikes.Targets = append(playlistLikes.Targets, queries.LikeTarget{
				ID:   createMusicId(),
				Type: "playlist",
			})
		}
	}

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(trackLikes)
	if err != nil {
		t.Error(err)
	}

	r := httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, int64(bulkSize/5))

	b, err = json.Marshal(playlistLikes)
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, int64(4*bulkSize/5))

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, bulkSize, int64(bulkSize), []queries.PartialLike{})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetType=track&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, bulkSize/5, int64(bulkSize/5), []queries.PartialLike{})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetType=playlist&limit=15000", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 4*bulkSize/5, int64(4*bulkSize/5), []queries.PartialLike{})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasLikes(t, r, 50, int64(bulkSize), []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: playlistLikes.Targets[0].ID,
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

	trackLikeTarget := queries.LikeTargetsFixed{
		Targets: queries.LikeTargets{
			{
				ID:   createMusicId(),
				Type: "track",
			},
		},
		Type: "track",
	}

	playlistLikeTarget := queries.LikeTargetsFixed{
		Targets: queries.LikeTargets{
			{
				ID:   createMusicId(),
				Type: "playlist",
			},
		},
		Type: "playlist",
	}

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(playlistLikeTarget)
	if err != nil {
		t.Error(err)
	}
	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 1)

	b, err = json.Marshal(trackLikeTarget)
	if err != nil {
		t.Error(err)
	}
	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token, 1)

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("POST", "/likes", bytes.NewReader(b))
	router.CommitMutating(t, r, Post, token2, 1)

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?targetIDs=%s", trackLikeTarget.Targets[0].ID), nil)
	testHasLikes(t, r, 2, 2, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: trackLikeTarget.Targets[0].ID,
	}, {
		EggsID:   os.Getenv("TESTUSER_ID2"),
		TargetID: trackLikeTarget.Targets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?targetIDs=%s&limit=1", trackLikeTarget.Targets[0].ID), nil)
	testHasLikes(t, r, 1, 2, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID2"),
		TargetID: trackLikeTarget.Targets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?targetIDs=%s", playlistLikeTarget.Targets[0].ID), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: playlistLikeTarget.Targets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetIDs=%s", os.Getenv("TESTUSER_ID"), trackLikeTarget.Targets[0].ID), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: trackLikeTarget.Targets[0].ID,
	}})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?eggsIDs=%s&targetType=%s", os.Getenv("TESTUSER_ID"), "track"), nil)
	testHasLikes(t, r, 1, 1, []queries.PartialLike{{
		EggsID:   os.Getenv("TESTUSER_ID"),
		TargetID: trackLikeTarget.Targets[0].ID,
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
