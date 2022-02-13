package routes

import (
	"hm2/players"

	"github.com/gorilla/mux"
)

func GetPlayerHandlers(r *mux.Router) {
	GetPlayerProfileHandlers(r)
}

func GetPlayerProfileHandlers(r *mux.Router) {
	r.HandleFunc("/api/player/{id:[0-9]+}", players.PlayerHandler).Methods("GET")
}
