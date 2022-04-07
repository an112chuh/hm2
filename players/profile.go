package players

import (
	"context"
	"errors"
	"hm2/config"
	"hm2/convert"
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
	ID, err := strconv.Atoi(id)
	if err != nil {
		res = result.SetErrorResult(`Неверный id игрока`)
		result.ReturnJSON(w, &res)
		return
	}
	res = GetPlayer(r, ID, user)
	result.ReturnJSON(w, &res)
}

func GetPlayer(r *http.Request, IDPlayer int, user config.User) (res result.ResultInfo) {
	var p Player
	ctx := r.Context()
	db := config.ConnectDB()
	p.Id = IDPlayer
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM list.players_list WHERE id = $1)`
	params := []interface{}{IDPlayer}
	err := db.QueryRowContext(ctx, query, params...).Scan(&exists)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, params...)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	if !exists {
		res = result.SetErrorResult(`Данного игрока не существует`)
		return
	}
	query = `SELECT team_id, name, surname, pos, nat, age, str, style, morale, readyness, tireness, price from list.players_list where id = $1`
	params = []interface{}{IDPlayer}
	err = db.QueryRowContext(ctx, query, params...).Scan(&p.TeamID, &p.Name, &p.Surname, &p.Pos, &p.Nat, &p.Age, &p.Str, &p.Style, &p.Morale, &p.Readyness, &p.Tireness, &p.Price)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, params...)
			res = result.SetErrorResult(report.UnknownError)
		}
		return
	}
	p.NatString, err = convert.NationToString(p.Nat)
	if err != nil {
		res = result.SetErrorResult(report.UnknownError)
		return
	}
	p.PosString = convert.PosToString(p.Pos)
	p.IsGK = false
	if p.Pos == 0 {
		p.IsGK = true
	}
	var gs GKSkills
	var s Skills
	if p.IsGK {
		query = `SELECT stick_handle, glove_handle, ricochet_control, fivehole, passing, reaction from players.gk_skills where player_id = $1`
		params = []interface{}{IDPlayer}
		err = db.QueryRowContext(ctx, query, params...).Scan(&gs.StickHandle, &gs.GloveHandle, &gs.RicochetContrl, &gs.FiveHole, &gs.Passing, &gs.Reaction)
		if err != nil {
			switch {
			case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
				res = result.SetErrorResult(report.CtxError)
			default:
				report.ErrorSQLServer(r, err, query, params...)
				res = result.SetErrorResult(report.UnknownError)
			}
			return
		}
		p.GKSkills = &gs
	} else {
		query = `SELECT speed, skating, slap_shot, wrist_shot, tackling, blocking, passing, vision, agressiveness, resistance, faceoff, side from players.skills where player_id = $1`
		params = []interface{}{IDPlayer}
		err = db.QueryRowContext(ctx, query, params...).Scan(&s.Speed, &s.Skating, &s.SlapShot, &s.WristShot, &s.Tackling, &s.Blocking, &s.Passing, &s.Vision, &s.Agressiveness, &s.Resistance, &s.Faceoff, &s.Hand)
		if err != nil {
			switch {
			case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
				res = result.SetErrorResult(report.CtxError)
			default:
				report.ErrorSQLServer(r, err, query, params...)
				res = result.SetErrorResult(report.UnknownError)
			}
			return
		}
		p.Skills = &s
	}
	query = `SELECT team_name, GP, G, A, P, PIM, PM, SOG, SOfG, rating from players.history where player_id = $1 ORDER BY id desc `
	params = []interface{}{IDPlayer}
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			res = result.SetErrorResult(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, params...)
			res = result.SetErrorResult(report.UnknownError)
		}
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
