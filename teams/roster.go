package teams

import (
	"context"
	"errors"
	"hm2/config"
	"hm2/convert"
	"hm2/get"
	"hm2/managers"
	"hm2/players"
	"hm2/report"
	"hm2/result"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Roster struct {
	ID        int                      `json:"id"`
	IsOwned   bool                     `json:"is_owned"`
	Name      string                   `json:"name"`
	Manager   string                   `json:"manager"`
	ManagerID int                      `json:"manager_id"`
	Country   string                   `json:"country"`
	CountryID int                      `json:"country_id"`
	City      string                   `json:"city"`
	Stadium   string                   `json:"stadium"`
	Capacity  int                      `json:"capacity"`
	Cash      int                      `json:"cash"`
	Players   []players.PlayerInRoster `json:"players"`
}

func RosterHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, false)
	vars := mux.Vars(r)
	id := vars["id"]
	ID, err := strconv.Atoi(id)
	if err != nil {
		res = result.SetErrorResult(`Неверный параметр id команды`)
		result.ReturnJSON(w, &res)
		return
	}
	res = GetRoster(r, ID, user)
	result.ReturnJSON(w, &res)
}

func RosterManagedHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, true)
	if !user.Authenticated {
		return
	}
	ctx := r.Context()
	ID, err := get.CurrentTeamByManager(ctx, user.ID)
	if err != nil {
		res = result.SetErrorResult(err.Error())
		result.ReturnJSON(w, &res)
		return
	}
	res = GetRoster(r, ID, user)
	result.ReturnJSON(w, &res)
}

func GetRoster(r *http.Request, IDTeam int, user config.User) (res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
	var roster Roster
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM list.team_list WHERE id = $1)`
	params := []interface{}{IDTeam}
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
		res = result.SetErrorResult(`Данной команды не существует`)
		return
	}
	query = `SELECT list.team_list.name, manager_id, country, city, stadium, capacity, cash from list.team_list
	INNER JOIN teams.data on teams.data.team_id = list.team_list.id WHERE list.team_list.id = $1`
	params = []interface{}{IDTeam}
	var IDManager int
	err = db.QueryRowContext(ctx, query, params...).Scan(&roster.Name, &IDManager, &roster.Country, &roster.City, &roster.Stadium, &roster.Capacity, &roster.Cash)
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
	if IDManager == user.ID {
		roster.IsOwned = true
	} else {
		roster.IsOwned = false
	}
	roster.ID = IDTeam
	roster.ManagerID = IDManager
	if IDManager > 0 {
		var name, surname string
		query = `SELECT name, surname from managers.data where id = $1`
		params = []interface{}{IDManager}
		err = db.QueryRowContext(ctx, query, params...).Scan(&name, &surname)
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
		roster.Manager = name + " " + surname
	}
	roster.CountryID, err = convert.NationToInt(roster.Country)
	if err != nil {
		res = result.SetErrorResult(report.UnknownError)
		return
	}
	var p players.PlayerInRoster
	query = `SELECT list.players_list.id, name,
	surname,
	pos, 
	nat,
	age,
	str,
	style,
	morale, 
	readyness,
	tireness,
	price,
	GP, G, A, P, PIM, PM, rating FROM list.players_list inner join players.history on list.players_list.id = players.history.id where list.players_list.team_id = $1`
	params = []interface{}{IDTeam}
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
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
	}
	defer rows.Close()
	for rows.Next() {
		var pos int
		err = rows.Scan(&p.ID, &p.Name, &p.Surname, &pos, &p.Nat, &p.Age, &p.Str, &p.Style, &p.Morale, &p.Readyness, &p.Tireness, &p.Price,
			&p.GP, &p.G, &p.A, &p.P, &p.PIM, &p.PM, &p.Rating)
		if err != nil {
			report.ErrorServer(r, err)
			res = result.SetErrorResult(report.UnknownError)
			return
		}
		p.PosString = convert.PosToString(pos)
		p.NatString, err = convert.NationToString(p.Nat)
		if err != nil {
			report.ErrorServer(r, err)
			res = result.SetErrorResult(report.UnknownError)
			return
		}
		roster.Players = append(roster.Players, p)
	}
	res.Done = true
	res.Items = roster
	return res
}
