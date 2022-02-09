package routes

import (
	"hm2/managers"
	"hm2/result"
	"net/http"

	"github.com/gorilla/mux"
)

func GetAllHandlers(r *mux.Router) {
	r.HandleFunc("/api/noauth", NoAuthHandler)
	GetManagerHandlers(r)
	GetTeamHandlers(r)
}

func NoAuthHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	res.Done = true
	IsLogged, user := managers.IsLogin(w, r, false)
	if IsLogged {
		res.Items = map[string]interface{}{"logged": "true", "id": user.ID}
	} else {
		res.Items = map[string]interface{}{"logged": "false"}
	}
	result.ReturnJSON(w, &res)
}
