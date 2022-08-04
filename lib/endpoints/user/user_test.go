package userendpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	userstubendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/userstub"
	"github.com/yayuyokitano/eggshellver/lib/queries"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

type TestUserResponse struct {
	EggsID         string    `json:"userName" db:"eggs_id"`
	DisplayName    string    `json:"displayName" db:"display_name"`
	IsArtist       bool      `json:"isArtist" db:"is_artist"`
	ImageDataPath  string    `json:"imageDataPath" db:"image_data_path"`
	PrefectureCode int       `json:"prefectureCode" db:"prefecture_code"`
	ProfileText    string    `json:"profile" db:"profile_text"`
	LastModified   time.Time `json:"lastModified" db:"last_modified"`
	Token          string    `json:"token" db:"token"`
}

func TestPost(t *testing.T) {
	services.Start()
	defer services.Stop()

	err := services.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer services.RollbackTransaction()

	w := httptest.NewRecorder()
	b, err := json.Marshal(Auth{
		Authorization: "Bearer ThisIsWrongToken",
		UserAgent:     os.Getenv("TESTUSER_USERAGENT"),
		ApVersion:     os.Getenv("TESTUSER_APVERSION"),
		DeviceID:      os.Getenv("TESTUSER_DEVICEID"),
		DeviceName:    os.Getenv("TESTUSER_DEVICENAME"),
	})
	r := httptest.NewRequest("POST", "/users", bytes.NewReader(b))
	Post(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code is %d, want %d", w.Code, http.StatusUnauthorized)
	}

	b, err = json.Marshal(Auth{
		Authorization: os.Getenv("TESTUSER_AUTHORIZATION"),
		UserAgent:     os.Getenv("TESTUSER_USERAGENT"),
		ApVersion:     os.Getenv("TESTUSER_APVERSION"),
		DeviceID:      os.Getenv("TESTUSER_DEVICEID"),
		DeviceName:    os.Getenv("TESTUSER_DEVICENAME"),
	})
	if err != nil {
		t.Error(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/users", bytes.NewReader(b))
	Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	token := w.Body.String()
	if token == "" {
		t.Errorf("Token is empty")
	}

	eggsID, err := queries.GetEggsIDByToken(context.Background(), token)
	if err != nil {
		t.Error(err)
	}
	if eggsID != os.Getenv("TESTUSER_ID") {
		t.Errorf("EggsID is %s, want %s", eggsID, os.Getenv("TESTUSER_ID"))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/users", bytes.NewReader(b))

	Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	if token != w.Body.String() {
		t.Errorf("Token is %s, want %s", token, w.Body.String())
	}

	token = w.Body.String()
	if token == "" {
		t.Errorf("Token is empty")
	}

	eggsID, err = queries.GetEggsIDByToken(context.Background(), token)
	if err != nil {
		t.Error(err)
	}

	if eggsID != os.Getenv("TESTUSER_ID") {
		t.Errorf("EggsID is %s, want %s", eggsID, os.Getenv("TESTUSER_ID"))
	}

	err = queries.UNSAFEDeleteUser(context.Background(), os.Getenv("TESTUSER_ID"))
	if err != nil {
		t.Error(err)
	}

	userStubs, err := json.Marshal([]queries.UserStub{{
		EggsID:         os.Getenv("TESTUSER_ID"),
		DisplayName:    "testuser",
		IsArtist:       false,
		ImageDataPath:  "https://example.com/testuser.png",
		PrefectureCode: 10,
		ProfileText:    "stan yayuyo",
	}})
	if err != nil {
		t.Error(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/userstubs", bytes.NewReader(userStubs))
	userstubendpoint.Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/users", bytes.NewReader(b))
	Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	token = w.Body.String()
	if token == "" {
		t.Errorf("Token is empty")
	}

	eggsID, err = queries.GetEggsIDByToken(context.Background(), token)
	if err != nil {
		t.Error(err)
	}

	if eggsID != os.Getenv("TESTUSER_ID") {
		t.Errorf("EggsID is %s, want %s", eggsID, os.Getenv("TESTUSER_ID"))
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

	createUser(t, os.Getenv("TESTUSER_AUTHORIZATION"))
	createUser(t, os.Getenv("TESTUSER_AUTHORIZATION2"))

	r := httptest.NewRequest("GET", fmt.Sprintf("/users?users=%s", os.Getenv("TESTUSER_ID")), nil)
	testHasUsers(t, r, 1, []string{os.Getenv("TESTUSER_ID")})

	r = httptest.NewRequest("GET", fmt.Sprintf("/users?users=%s,%s", os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2")), nil)
	testHasUsers(t, r, 2, []string{os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2")})

	r = httptest.NewRequest("GET", fmt.Sprintf("/users?users=%s,%s,%s", os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2"), os.Getenv("TESTUSER_FAILID")), nil)
	testHasUsers(t, r, 2, []string{os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2")})

	w := httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/users?users=%s", os.Getenv("TESTUSER_FAILID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	if w.Body.String() != "[]" {
		t.Errorf("Body is %s, want %s", w.Body.String(), "[]")
	}
}

func testHasUsers(t *testing.T, r *http.Request, num int, expectedUsers []string) {
	t.Helper()
	w := httptest.NewRecorder()
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d, body %s", w.Code, http.StatusOK, w.Body.String())
	}

	var users queries.UserStubs
	err := json.Unmarshal([]byte(w.Body.String()), &users)
	if err != nil {
		t.Error(err)
	}
	if len(users) != num {
		t.Errorf("Returned %d users, want %d", len(users), num)
	}
	for _, user := range expectedUsers {
		if !users.ContainsID(user) {
			t.Errorf("Expected slice to include user %v", user)
		}
	}
}

func createUser(t *testing.T, authorization string) {
	w := httptest.NewRecorder()
	b, err := json.Marshal(Auth{
		Authorization: authorization,
		UserAgent:     os.Getenv("TESTUSER_USERAGENT"),
		ApVersion:     os.Getenv("TESTUSER_APVERSION"),
		DeviceID:      os.Getenv("TESTUSER_DEVICEID"),
		DeviceName:    os.Getenv("TESTUSER_DEVICENAME"),
	})
	r := httptest.NewRequest("POST", "/users", bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
}
