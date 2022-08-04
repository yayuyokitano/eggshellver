package userstubendpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	userendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/user"
	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

func TestPost(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	testUsers := []queries.UserStub{
		{
			EggsID:         os.Getenv("TESTUSER_ID"),
			DisplayName:    "testuser",
			IsArtist:       false,
			ImageDataPath:  "https://example.com/testuser.png",
			PrefectureCode: 10,
			ProfileText:    "stan yayuyo",
		},
		{
			EggsID:         os.Getenv("TESTUSER_ID2"),
			DisplayName:    "testuser numero dos",
			IsArtist:       true,
			ImageDataPath:  "",
			PrefectureCode: 28,
			ProfileText:    "",
		},
	}

	b, err := json.Marshal(testUsers)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/userstubs", bytes.NewReader(b))
	Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "2 users inserted" {
		t.Errorf("expected body %s, got %s", "2 users inserted", w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/users?users=%s,%s", os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2")), nil)
	userendpoint.Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d, Body %s", http.StatusOK, w.Code, w.Body.String())
	}
	var users queries.UserStubs
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Error(err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
	if !users.Contains(testUsers[0]) {
		t.Errorf("expected users to contain %v, got %v", testUsers[0], users)
	}
	if !users.Contains(testUsers[1]) {
		t.Errorf("expected users to contain %v, got %v", testUsers[1], users)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/userstubs", strings.NewReader(""))
	Post(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d, Body %s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/userstubs", strings.NewReader("{}"))
	Post(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d, Body %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}
