package routes

import (
	"hm2/teams"

	"github.com/gorilla/mux"
)

func GetTeamHandlers(r *mux.Router) {
	GetCreateTeamHandlers(r)
	GetRosterHandlers(r)
	GetFreeTeamsHandlers(r)
	GetBuyTeamsHandlers(r)
}

func GetCreateTeamHandlers(r *mux.Router) {
	r.HandleFunc("/adm/api/create_team", teams.CreateTeamHandler).Methods("GET")
	r.HandleFunc("/adm/api/create_team", teams.CreateTeamConfirmHandler).Methods("POST")
}

func GetRosterHandlers(r *mux.Router) {
	r.HandleFunc("/api/roster/{id:[0-9]+}", teams.RosterHandler).Methods("GET")
	r.HandleFunc("/api/roster", teams.RosterManagedHandler).Methods("GET")
}

func GetFreeTeamsHandlers(r *mux.Router) {
	r.HandleFunc("/api/freeteam", teams.FreeTeamsListHandler).Methods("GET")
}

func GetBuyTeamsHandlers(r *mux.Router) {
	r.HandleFunc("/api/roster/{id:[0-9]+}/buy", teams.BuyTeamHandler).Methods("GET")
	r.HandleFunc("/api/roster/{id:[0-9]+}/sell", teams.SellTeamHandler).Methods("GET")
	r.HandleFunc("/api/roster/{id:[0-9]+}/auc", teams.AucTeamHandler).Methods("GET")
	r.HandleFunc("/adm/api/team/edit", teams.EditTeamsHandler).Methods("GET")
	r.HandleFunc("/adm/api/team/edit", teams.EditTeamsConfirmHandler).Methods("PUT")
}
