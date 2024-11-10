package cookies

import (
	"encoding/json"
	"github.com/gorilla/sessions"
	"net/http"
	"os"
	"time"
)

var Store *sessions.CookieStore

func SetNewCookieStore() {
	Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_TOKEN")))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		Secure:   false,
		HttpOnly: true,
	}
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "user-session")

	session.Options.MaxAge = -1
	session.Save(r, w) // nolint:errcheck
}

func Serialize(cookies []*http.Cookie) (string, error) {
	data, err := json.Marshal(cookies)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Deserialize(data string) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	err := json.Unmarshal([]byte(data), &cookies)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}

func WriteToFile(filename, data string) error {
	return os.WriteFile(filename, []byte(data), 0600)
}

func ReadFromFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
