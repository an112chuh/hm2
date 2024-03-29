package routes

import (
	"hm2/config"
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
	GetStaticHandlers(r)
}

func NoAuthHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	res.Done = true
	user := managers.IsLogin(w, r, false)
	if user.Authenticated {
		var cash int
		db := config.ConnectDB()
		query := `SELECT cash FROM managers.data WHERE id = $1`
		params := []interface{}{user.ID}
		_ = db.QueryRowContext(r.Context(), query, params...).Scan(&cash)
		res.Items = map[string]interface{}{"logged": "true", "id": user.ID, "nickname": user.Username, "cash": cash}
	} else {
		res.Items = map[string]interface{}{"logged": "false"}
	}
	result.ReturnJSON(w, &res)
}
