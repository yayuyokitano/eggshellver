package userendpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/yayuyokitano/eggshellver/lib/router"
)

func CreateTestUser(userNum int) (token string, err error) {
	w := httptest.NewRecorder()
	var b []byte
	switch userNum {
	case 1:
		b, err = json.Marshal(Auth{
			Authorization: os.Getenv("TESTUSER_AUTHORIZATION"),
			UserAgent:     os.Getenv("TESTUSER_USERAGENT"),
			ApVersion:     os.Getenv("TESTUSER_APVERSION"),
			DeviceID:      os.Getenv("TESTUSER_DEVICEID"),
			DeviceName:    os.Getenv("TESTUSER_DEVICENAME"),
		})
	case 2:
		b, err = json.Marshal(Auth{
			Authorization: os.Getenv("TESTUSER_AUTHORIZATION2"),
			UserAgent:     os.Getenv("TESTUSER_USERAGENT"),
			ApVersion:     os.Getenv("TESTUSER_APVERSION"),
			DeviceID:      os.Getenv("TESTUSER_DEVICEID2"),
			DeviceName:    os.Getenv("TESTUSER_DEVICENAME"),
		})
	default:
		err = fmt.Errorf("invalid user number %d", userNum)
		return
	}

	r := httptest.NewRequest("POST", "/users", bytes.NewReader(b))
	if err != nil {
		return
	}
	router.HandleMethod(Post, w, r)
	if w.Code != http.StatusOK {
		err = fmt.Errorf("status code is %d, want %d. Body %s", w.Code, http.StatusOK, w.Body.String())
		return
	}
	str := w.Body.String()
	token = str[1 : len(str)-1]
	if token == "" {
		err = fmt.Errorf("token is empty")
		return
	}
	return
}
