package managers

import (
	"hm2/config"
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func RegManagerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "cookie-name")
	if err != nil {

	}
}
