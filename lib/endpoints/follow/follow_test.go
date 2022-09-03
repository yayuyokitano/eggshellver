package followendpoint

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
	userstubendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/userstub"
	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/router"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

var testUserStubs = []queries.UserStub{
	{
		EggsID:         "1",
		DisplayName:    "1",
		IsArtist:       true,
		ImageDataPath:  "datapath",
		PrefectureCode: 4,
		ProfileText:    "profiletext",
	},
	{
		EggsID:         "2",
		DisplayName:    "2",
		IsArtist:       false,
		ImageDataPath:  "",
		PrefectureCode: 0,
		ProfileText:    "",
	},
	{
		EggsID:         "3",
		DisplayName:    "3",
		IsArtist:       false,
		ImageDataPath:  "",
		PrefectureCode: 0,
		ProfileText:    "",
	},
	{
		EggsID:         "4",
		DisplayName:    "4",
		IsArtist:       false,
		ImageDataPath:  "",
		PrefectureCode: 0,
		ProfileText:    "",
	},
	{
		EggsID:         "5",
		DisplayName:    "5",
		IsArtist:       false,
		ImageDataPath:  "",
		PrefectureCode: 0,
		ProfileText:    "",
	},
}

func TestInit(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}
	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID2"))
	if err != nil {
		t.Error(err)
	}

	err = services.CommitTransaction()
	if err != nil {
		t.Fatal(err)
	}
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

	initUserStubs(t)

	r := httptest.NewRequest("POST", "/follows", strings.NewReader(`["1"]`))
	router.CommitMutating(t, r, Post, token, 1)

	r = httptest.NewRequest("POST", "/follows", strings.NewReader(`["1"]`))
	router.CommitMutating(t, r, Post, token, 0)

}

func TestPost(t *testing.T) {
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

	initUserStubs(t)

	r := httptest.NewRequest("POST", "/follows", strings.NewReader(`["1","2","3","4","5"]`))
	router.CommitMutating(t, r, Post, token, 5)

	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasFollowersFollowees(t, r, 5, 5, []string{os.Getenv("TESTUSER_ID")}, []string{"1", "2", "3", "4", "5"})

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasFollowersFollowees(t, r, 0, 0, []string{}, []string{})

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
	r := httptest.NewRequest("GET", "/follows", nil)
	router.HandleMethod(Get, w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	router.HandleMethod(Get, w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != `{"follows":[],"total":0}` {
		t.Errorf("Body is %s, want %s", w.Body.String(), `{"follows":[],"total":0}`)
	}

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	initUserStubs(t)

	r = httptest.NewRequest("POST", "/follows", strings.NewReader(`["1","2"]`))
	router.CommitMutating(t, r, Post, token, 2)

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest("POST", "/follows", strings.NewReader(`["1"]`))
	router.CommitMutating(t, r, Post, token2, 1)

	followeeIDs := []string{"1", "2"}

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?followeeIDs=%s", followeeIDs[0]), nil)
	testHasFollowersFollowees(t, r, 2, 2, []string{os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2")}, []string{followeeIDs[0]})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?followeeIDs=%s&limit=1", followeeIDs[0]), nil)
	testHasFollowersFollowees(t, r, 1, 2, []string{os.Getenv("TESTUSER_ID2")}, []string{followeeIDs[0]})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?followeeIDs=%s", followeeIDs[1]), nil)
	testHasFollowersFollowees(t, r, 1, 1, []string{os.Getenv("TESTUSER_ID")}, []string{followeeIDs[1]})

	r = httptest.NewRequest("GET", fmt.Sprintf("/likes?followerIDs=%s&followeeIDs=%s", os.Getenv("TESTUSER_ID"), followeeIDs[0]), nil)
	testHasFollowersFollowees(t, r, 1, 1, []string{os.Getenv("TESTUSER_ID")}, []string{followeeIDs[0]})

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

	initUserStubs(t)

	r := httptest.NewRequest("PUT", "/follows", strings.NewReader(`["1","2","3"]`))
	router.CommitMutating(t, r, Put, token, 3)

	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasFollowersFollowees(t, r, 3, 3, []string{os.Getenv("TESTUSER_ID")}, []string{"1", "2", "3"})

	r = httptest.NewRequest("PUT", "/follows", strings.NewReader(`["3","4","5"]`))
	router.CommitMutating(t, r, Put, token, 4)

	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasFollowersFollowees(t, r, 4, 4, []string{os.Getenv("TESTUSER_ID")}, []string{"2", "3", "4", "5"})

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	router.HandleMethod(Get, w, r)
	var follows queries.StructuredFollows
	err = json.Unmarshal(w.Body.Bytes(), &follows)
	if err != nil {
		t.Error(err)
	}
	//Timestamps should not have changed for the follows that had been recorded
	if follows.Follows[0].Followee.EggsID != "4" {
		t.Errorf("Expected followee to be 4, got %s", follows.Follows[0].Followee.EggsID)
	}
	if follows.Follows[2].Followee.EggsID != "2" {
		t.Errorf("Expected followee to be 2, got %s", follows.Follows[2].Followee.EggsID)
	}
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

	initUserStubs(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/follow", nil)
	router.HandleMethod(Toggle, w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code to be %d, got %d", http.StatusBadRequest, w.Code)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/follow/", nil)
	router.HandleMethod(Toggle, w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code to be %d, got %d", http.StatusBadRequest, w.Code)
	}

	r = httptest.NewRequest("POST", fmt.Sprintf("/follow/%s", testUserStubs[0].EggsID), nil)
	router.CommitToggling(t, r, Toggle, token, true)

	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasFollowersFollowees(t, r, 1, 1, []string{os.Getenv("TESTUSER_ID")}, []string{testUserStubs[0].EggsID})

	r = httptest.NewRequest("POST", fmt.Sprintf("/follow/%s", testUserStubs[0].EggsID), nil)
	router.CommitToggling(t, r, Toggle, token, false)

	r = httptest.NewRequest("GET", fmt.Sprintf("/follows?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasFollowersFollowees(t, r, 0, 0, []string{}, []string{})

}

func testHasFollowersFollowees(t *testing.T, r *http.Request, num int, total int64, followerIDs []string, followeeIDs []string) {
	t.Helper()
	w := httptest.NewRecorder()
	router.HandleMethod(Get, w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}

	var follows queries.StructuredFollows
	err := json.Unmarshal(w.Body.Bytes(), &follows)
	if err != nil {
		t.Error(err)
	}
	if len(follows.Follows) != num {
		t.Errorf("Returned %d follows, want %d", len(follows.Follows), num)
	}
	if follows.Total != total {
		t.Errorf("Returned %d total, want %d", follows.Total, total)
	}
	for _, followerID := range followerIDs {
		if !follows.ContainsFollowerID(followerID) {
			t.Errorf("Expected slice to include follower %s", followerID)
		}
	}
	for _, followeeID := range followeeIDs {
		if !follows.ContainsFolloweeID(followeeID) {
			t.Errorf("Expected slice to include followee %s", followeeID)
		}
	}
}

func initUserStubs(t *testing.T) {
	t.Helper()
	b, err := json.Marshal(testUserStubs)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/userstubs", bytes.NewReader(b))
	router.HandleMethod(userstubendpoint.Post, w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
}
