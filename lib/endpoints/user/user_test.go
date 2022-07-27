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
	r := httptest.NewRequest("POST", "/user", bytes.NewReader(b))
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
	r = httptest.NewRequest("POST", "/user", bytes.NewReader(b))
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
	r = httptest.NewRequest("POST", "/user", bytes.NewReader(b))

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
	r = httptest.NewRequest("POST", "/userstub", bytes.NewReader(userStubs))
	userstubendpoint.Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/user", bytes.NewReader(b))
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

	w := httptest.NewRecorder()
	b, err := json.Marshal(Auth{
		Authorization: os.Getenv("TESTUSER_AUTHORIZATION"),
		UserAgent:     os.Getenv("TESTUSER_USERAGENT"),
		ApVersion:     os.Getenv("TESTUSER_APVERSION"),
		DeviceID:      os.Getenv("TESTUSER_DEVICEID"),
		DeviceName:    os.Getenv("TESTUSER_DEVICENAME"),
	})
	r := httptest.NewRequest("POST", "/user", bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	w = httptest.NewRecorder()
	b, err = json.Marshal(Auth{
		Authorization: os.Getenv("TESTUSER_AUTHORIZATION2"),
		UserAgent:     os.Getenv("TESTUSER_USERAGENT"),
		ApVersion:     os.Getenv("TESTUSER_APVERSION"),
		DeviceID:      os.Getenv("TESTUSER_DEVICEID2"),
		DeviceName:    os.Getenv("TESTUSER_DEVICENAME"),
	})
	r = httptest.NewRequest("POST", "/user", bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	Post(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/user?users=%s", os.Getenv("TESTUSER_ID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	var users []queries.UserStub
	err = json.NewDecoder(w.Body).Decode(&users)
	if len(users) != 1 {
		t.Errorf("Users count is %d, want %d", len(users), 1)
	}
	if users[0].EggsID != os.Getenv("TESTUSER_ID") {
		t.Errorf("User ID is %s, want %s", users[0].EggsID, os.Getenv("TESTUSER_ID"))
	}
	if users[0].IsArtist != false {
		t.Errorf("User isArtist is %t, want %t", users[0].IsArtist, false)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/user?users=%s,%s", os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	err = json.NewDecoder(w.Body).Decode(&users)
	if len(users) != 2 {
		t.Errorf("Users count is %d, want %d", len(users), 2)
	}
	if users[0].EggsID != os.Getenv("TESTUSER_ID") || users[1].EggsID != os.Getenv("TESTUSER_ID2") {
		t.Errorf("User IDs are %s,%s, want %s,%s", users[0].EggsID, users[1].EggsID, os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2"))
	}
	if users[0].IsArtist != false || users[1].IsArtist != false {
		t.Errorf("User isArtist is %t,%t, want %t,%t", users[0].IsArtist, users[1].IsArtist, false, false)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/user?users=%s,%s,%s", os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2"), os.Getenv("TESTUSER_FAILID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}
	err = json.NewDecoder(w.Body).Decode(&users)
	if len(users) != 2 {
		t.Errorf("Users count is %d, want %d", len(users), 2)
	}
	if users[0].EggsID != os.Getenv("TESTUSER_ID") || users[1].EggsID != os.Getenv("TESTUSER_ID2") {
		t.Errorf("User IDs are %s,%s, want %s,%s", users[0].EggsID, users[1].EggsID, os.Getenv("TESTUSER_ID"), os.Getenv("TESTUSER_ID2"))
	}
	if users[0].IsArtist != false || users[1].IsArtist != false {
		t.Errorf("User isArtist is %t,%t, want %t,%t", users[0].IsArtist, users[1].IsArtist, false, false)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", fmt.Sprintf("/user?users=%s", os.Getenv("TESTUSER_FAILID")), nil)
	Get(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
	}

	if w.Body.String() != "[]" {
		t.Errorf("Body is %s, want %s", w.Body.String(), "[]")
	}
}
