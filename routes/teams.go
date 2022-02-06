package routes

import (
	"hm2/teams"

	"github.com/gorilla/mux"
)

func GetTeamHandlers(r *mux.Router) {
	GetCreateTeamHandlers(r)
}

func GetCreateTeamHandlers(r *mux.Router) {
	r.HandleFunc("/adm/api/create_team", teams.CreateTeamHandler).Methods("GET")
}
