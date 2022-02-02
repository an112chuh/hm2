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
	r.HandleFunc("/api/login", managers.Login).Methods("POST")
}

func GetManagerProfileHandlers(r *mux.Router) {

}
