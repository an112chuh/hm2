package routes

import (
	"hm2/managers"

	"github.com/gorilla/mux"
)

func GetManagerHandlers(r *mux.Router) {
	GetManagerChangeHandlers(r)
	GetManagerAuthHandlers(r)
	GetManagerProfileHandlers(r)
}

func GetManagerChangeHandlers(r *mux.Router) {
	r.HandleFunc("/api/register", managers.RegManagerHandler).Methods("POST")
	r.HandleFunc("/api/delete", managers.DeleteManagerHandler)

}

func GetManagerAuthHandlers(r *mux.Router) {
	r.HandleFunc("/api/login", managers.LoginHandler).Methods("POST")
	r.HandleFunc("/api/logout", managers.DeleteManagerHandler)
}

func GetManagerProfileHandlers(r *mux.Router) {
	r.HandleFunc("/api/profile/{id:[0-9]+}", managers.ProfileHandler).Methods("GET")
	r.HandleFunc("/api/profile/{id:[0-9]+}", managers.EditProfileHandler).Methods("PUT")
}
