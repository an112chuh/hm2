package routes

import (
	"hm2/result"
	"net/http"

	"github.com/gorilla/mux"
)

func GetAllHandlers(r *mux.Router) {
	r.HandleFunc("/api/noauth", NoAuthHandler)
	GetManagerHandlers(r)
}

func NoAuthHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	res.Done = true
	res.Items = "Hello, world!"
	result.ReturnJSON(w, &res)
}
