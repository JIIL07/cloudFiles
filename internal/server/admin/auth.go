package admin

import (
	"encoding/json"
	"github.com/JIIL07/jcloud/internal/server/cookies"
	"github.com/JIIL07/jcloud/internal/server/storage"
	j "github.com/JIIL07/jcloud/pkg/json"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only method GET is allowed")) // nolint:errcheck
		return
	}

	a := r.URL.Query().Get("admin")

	d := os.Getenv("ADMIN_USER")

	var u storage.User
	err := json.Unmarshal([]byte(d), &u)
	if err != nil {
		http.Error(w, "Invalid admin user configuration", http.StatusInternalServerError)
		return
	}

	if a == u.Username {
		session, err := cookies.Store.Get(r, "admin")
		if err != nil {
			j.RespondWithError(w, err)
			return
		}

		session.Values["admin"] = true
		session.Values["sql"] = true
		session.Values["cmd"] = true

		err = sessions.Save(r, w)
		if err != nil {
			j.RespondWithError(w, err)
			return
		}

		w.Write([]byte("Session established")) // nolint:errcheck
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized")) // nolint:errcheck
}

func CheckHandler(w http.ResponseWriter, r *http.Request) {
	s, err := cookies.Store.Get(r, "admin")
	if err != nil {
		j.RespondWithError(w, err)
		return
	}

	if s.IsNew {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized")) // nolint:errcheck
		return
	}

	w.Write([]byte("Admin authorized")) // nolint:errcheck
}
