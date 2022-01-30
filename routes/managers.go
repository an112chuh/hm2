package routes

import (
	"hm2/managers"

	"github.com/gorilla/mux"
)

func GetManagerHandlers(r *mux.Router) {
	GetManagerAuthHandlers(r)
}

func GetManagerAuthHandlers(r *mux.Router) {
	r.HandleFunc("/api/login", managers.Login)
}
