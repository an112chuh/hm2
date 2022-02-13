package players

import (
	"hm2/config"
	"hm2/managers"
	"hm2/report"
	"hm2/result"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PlayerInRoster struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Surname   string  `json:"surname"`
	PosString string  `json:"pos_string"`
	Nat       int     `json:"nat"`
	NatString string  `json:"nat_string"`
	Age       int     `json:"age"`
	Str       int     `json:"str"`
	Readyness int     `json:"readyness"`
	Morale    int     `json:"morale"`
	Tireness  int     `json:"tireness"`
	Style     int     `json:"style"`
	GP        int     `json:"GP"`
	G         int     `json:"G"`
	A         int     `json:"A"`
	P         int     `json:"P"`
	PIM       int     `json:"PIM"`
	PM        int     `json:"PM"`
	Rating    float64 `json:"rating"`
	Price     int     `json:"price"`
}

func PlayerHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, false)
	vars := mux.Vars(r)
	id := vars["id"]
	ID, _ := strconv.Atoi(id)
	res = GetPlayer(r, ID, user)
	result.ReturnJSON(w, &res)
}

func GetPlayer(r *http.Request, IDPlayer int, user config.User) (res result.ResultInfo) {
	var p Player
	db := config.ConnectDB()
	p.Id = IDPlayer
	query := `SELECT team_id, name, surname, pos, nat, age, str, style, morale, readyness, tireness, price from list.players_list where id = $1`
	params := []interface{}{IDPlayer}
	err := db.QueryRow(query, params...).Scan(&p.TeamID, &p.Name, &p.Surname, &p.Pos, &p.Nat, &p.Age, &p.Str, &p.Style, &p.Morale, &p.Readyness, &p.Tireness, &p.Price)
	if err != nil {
		res = result.SetErrorResult(report.UnknownError)
		report.ErrorSQLServer(r, err, query, params...)
		return
	}
	p.NatString, err = ConvertNationToString(p.Nat)
	if err != nil {
		res = result.SetErrorResult(report.UnknownError)
		report.ErrorSQLServer(r, err, query, params...)
		return
	}
	p.PosString = ConvertPosToString(p.Pos)
	p.IsGK = false
	if p.Pos == 0 {
		p.IsGK = true
	}
	if p.IsGK {
		query = `SELECT stick_handle, glove_handle, ricochet_control, fivehole, passing, reaction from players.gk_skills where player_id = $1`
		params = []interface{}{IDPlayer}
		err = db.QueryRow(query, params...).Scan(&p.GKSkills.StickHandle, &p.GKSkills.GloveHandle, &p.GKSkills.RicochetContrl, &p.GKSkills.FiveHole, &p.GKSkills.Passing, &p.GKSkills.Reaction)
		if err != nil {
			res = result.SetErrorResult(report.UnknownError)
			report.ErrorSQLServer(r, err, query, params...)
			return
		}
	} else {
		query = `SELECT speed, skating, slap_shot, wrist_shot, tackling, blocking, passing, vision, agressiveness, resistance, faceoff, side from players.skills where player_id = $1`
		params = []interface{}{IDPlayer}
		err = db.QueryRow(query, params...).Scan(&p.Skills.Speed, &p.Skills.Skating, &p.Skills.SlapShot, &p.Skills.WristShot, &p.Skills.Tackling, &p.Skills.Blocking, &p.Skills.Passing, &p.Skills.Vision, &p.Skills.Agressiveness, &p.Skills.Resistance, &p.Skills.Faceoff, &p.Skills.Hand)
		if err != nil {
			res = result.SetErrorResult(report.UnknownError)
			report.ErrorSQLServer(r, err, query, params...)
			return
		}
	}
	query = `SELECT team_name, GP, G, A, P, PIM, PM, SOG, SOfG, rating from players.history where player_id = $1 ORDER BY id desc `
	params = []interface{}{IDPlayer}
	rows, err := db.Query(query, params...)
	if err != nil {
		res = result.SetErrorResult(report.UnknownError)
		report.ErrorSQLServer(r, err, query, params...)
		return
	}
	ctr := 0
	for rows.Next() {
		var h History
		err = rows.Scan(&h.Team, &h.GP, &h.G, &h.A, &h.P, &h.PIM, &h.PM, &h.SOG, &h.SOfG, &h.Rating)
		if err != nil {
			res = result.SetErrorResult(report.UnknownError)
			report.ErrorSQLServer(r, err, query, params...)
			return
		}
		if ctr == 0 {
			p.TeamName = h.Team
		}
		p.History = append(p.History, h)
	}
	res.Done = true
	res.Items = p
	return res
}