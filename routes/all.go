package routes

import "github.com/gorilla/mux"

func GetAllHandlers(r *mux.Router) {
	GetManagerHandlers(r)
}
