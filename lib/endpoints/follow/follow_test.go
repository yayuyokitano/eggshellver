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

	b, err := json.Marshal(testUserStubs)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/userstub", bytes.NewReader(b))
	userstubendpoint.Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/follow", strings.NewReader(`["1","2","3","4","5"]`))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/follow?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}

	var follows queries.StructuredFollows
	err = json.Unmarshal([]byte(w.Body.String()), &follows)
	if err != nil {
		t.Error(err)
	}
	if len(follows.Follows) != 5 {
		t.Errorf("Returned %d liked tracks, want %d", len(follows.Follows), 5)
	}
	if follows.Total != 5 {
		t.Errorf("Returned %d total, want %d", follows.Total, 5)
	}

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/follow?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var emptyFollows queries.StructuredFollows
	err = json.Unmarshal([]byte(w.Body.String()), &emptyFollows)
	if err != nil {
		t.Error(err)
	}
	if len(emptyFollows.Follows) != 0 {
		t.Errorf("Returned %d liked tracks, want %d", len(emptyFollows.Follows), 0)
	}
	if emptyFollows.Total != 0 {
		t.Errorf("Returned %d total, want %d", emptyFollows.Total, 0)
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
	r := httptest.NewRequest("GET", "/follow", nil)
	Get(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/follow?followerIDs=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != `{"follows":[],"total":0}` {
		t.Errorf("Body is %s, want %s", w.Body.String(), `{"follows":[],"total":0}`)
	}

	followeeIDs := []string{"1", "2"}

	token, err := userendpoint.CreateTestUser(1)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(testUserStubs)

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/userstub", bytes.NewReader(b))
	userstubendpoint.Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	b, err = json.Marshal(followeeIDs)
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/follow", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != fmt.Sprintf("%s followed 2 users", os.Getenv("TESTUSER_ID")) {
		t.Errorf("Body is %s, want %s", w.Body.String(), fmt.Sprintf("%s followed 2 users", os.Getenv("TESTUSER_ID")))
	}

	token2, err := userendpoint.CreateTestUser(2)
	if err != nil {
		t.Error(err)
	}
	b, err = json.Marshal(followeeIDs[:1])
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/follow", bytes.NewReader(b))
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token2))
	Post(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != fmt.Sprintf("%s followed 1 users", os.Getenv("TESTUSER_ID2")) {
		t.Errorf("Body is %s, want %s", w.Body.String(), fmt.Sprintf("%s followed 1 users", os.Getenv("TESTUSER_ID2")))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?followeeIDs=%s", followeeIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}

	var follows queries.StructuredFollows
	err = json.Unmarshal([]byte(w.Body.String()), &follows)
	if err != nil {
		t.Error(err)
	}
	if len(follows.Follows) != 2 {
		t.Errorf("Returned %d following users, want %d", len(follows.Follows), 2)
	}
	if follows.Total != 2 {
		t.Errorf("Returned %d total following users, want %d", follows.Total, 2)
	}
	if !follows.ContainsFollowerID(os.Getenv("TESTUSER_ID")) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}
	if !follows.ContainsFollowerID(os.Getenv("TESTUSER_ID2")) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID2"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?followeeIDs=%s&limit=1", followeeIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var limitedFollows queries.StructuredFollows
	err = json.Unmarshal([]byte(w.Body.String()), &limitedFollows)
	if err != nil {
		t.Error(err)
	}
	if len(limitedFollows.Follows) != 1 {
		t.Errorf("Returned %d following users, want %d", len(limitedFollows.Follows), 1)
	}
	if limitedFollows.Total != 2 {
		t.Errorf("Returned %d total following users, want %d", limitedFollows.Total, 2)
	}
	if !limitedFollows.ContainsFollowerID(os.Getenv("TESTUSER_ID")) && !limitedFollows.ContainsFollowerID(os.Getenv("TESTUSER_ID2")) {
		t.Errorf("Expected slice to include %s or %s", os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?followeeIDs=%s", followeeIDs[1]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var follows2 queries.StructuredFollows
	err = json.Unmarshal([]byte(w.Body.String()), &follows2)
	if err != nil {
		t.Error(err)
	}
	if len(follows2.Follows) != 1 {
		t.Errorf("Returned %d liking users, want %d", len(follows2.Follows), 1)
	}
	if follows2.Total != 1 {
		t.Errorf("Returned %d total liking users, want %d", follows2.Total, 1)
	}
	if !follows2.ContainsFollowerID(os.Getenv("TESTUSER_ID")) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/like?followerIDs=%s&followeeIDs=%s", os.Getenv("TESTUSER_ID"), followeeIDs[0]), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() == "" {
		t.Errorf("Body is empty")
	}
	var specifiedFollows queries.StructuredFollows
	err = json.Unmarshal([]byte(w.Body.String()), &specifiedFollows)
	if err != nil {
		t.Error(err)
	}
	if len(specifiedFollows.Follows) != 1 {
		t.Errorf("Returned %d likes, want %d", len(specifiedFollows.Follows), 1)
	}
	if specifiedFollows.Total != 1 {
		t.Errorf("Returned %d total likes, want %d", specifiedFollows.Total, 1)
	}
	if !specifiedFollows.ContainsFollowerID(os.Getenv("TESTUSER_ID")) || !specifiedFollows.ContainsFolloweeID(followeeIDs[0]) {
		t.Errorf("Expected slice to include %s", os.Getenv("TESTUSER_ID"))
	}

}
