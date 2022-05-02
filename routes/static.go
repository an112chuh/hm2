package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

func GetStaticHandlers(r *mux.Router) {
	r.StrictSlash(true)
	staticDir := `/public/`
	r.PathPrefix(staticDir).Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("."+staticDir))))
}
