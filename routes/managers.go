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
	r.HandleFunc("/api/profile", managers.ProfileHandler)
	r.HandleFunc("/api/profile/{id:[0-9]+}", managers.IDProfileHandler)
	r.HandleFunc("/api/profile/edit", managers.EditProfileHandler).Methods("GET")
	r.HandleFunc("/api/profile/edit", managers.EditProfileConfirmHandler).Methods("PUT")
}
