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
	GetPlayerHandlers(r)
}

func NoAuthHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	res.Done = true
	user := managers.IsLogin(w, r, false)
	if user.Authenticated {
		res.Items = map[string]interface{}{"logged": "true", "id": user.ID}
	} else {
		res.Items = map[string]interface{}{"logged": "false"}
	}
	result.ReturnJSON(w, &res)
}
