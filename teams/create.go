package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hm2/bases"
	"hm2/config"
	"hm2/managers"
	"hm2/players"
	"hm2/report"
	"hm2/result"
	"io"
	"net/http"
)

type CreateTeamData struct {
	Name    string `json:"team_name"`
	Country string `json:"country"`
	City    string `json:"city"`
	Stadium string `json:"stadium"`
}

func CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, true)
	if !user.Authenticated {
		res = result.SetErrorResult(`У вас нет прав для доступа к данной страницы`)
		result.ReturnJSON(w, &res)
		return
	}
	if user.Rights != config.Admin {
		res = result.SetErrorResult(`У вас нет прав для доступа к данной страницы`)
		result.ReturnJSON(w, &res)
		return
	}
	res.Done = true
	res.Items = user.ID
	result.ReturnJSON(w, &res)
}

func CreateTeamConfirmHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, true)
	if !user.Authenticated {
		res = result.SetErrorResult(`У вас нет прав для доступа к данной страницы`)
		result.ReturnJSON(w, &res)
		return
	}
	if user.Rights != config.Admin {
		res = result.SetErrorResult(`У вас нет прав для совершения данной операции`)
		result.ReturnJSON(w, &res)
		return
	}
	var data CreateTeamData
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err := json.Unmarshal(b, &data)
	fmt.Println(data)
	if err != nil {
		report.ErrorServer(r, err)
	}
	res = CreateTeamConfirm(r, data)
	result.ReturnJSON(w, &res)
}

func CreateTeamConfirm(r *http.Request, data CreateTeamData) (res result.ResultInfo) {
	db := config.ConnectDB()
	tx, err := db.Begin()
	if err != nil {
		report.ErrorServer(r, err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	ctx := r.Context()
	IsCountryExist := CheckCountry(r, data.Country)
	if !IsCountryExist {
		res = result.SetErrorResult(`Данной страны не существует в списке стран`)
		return
	}
	var IDTeam int
	query := `INSERT into list.team_list (name, country, city, stadium, capacity, manager_id) VALUES ($1, $2, $3, $4, 100, -1) RETURNING id`
	params := []interface{}{data.Name, data.Country, data.City, data.Stadium}
	err = tx.QueryRowContext(ctx, query, params...).Scan(&IDTeam)
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
	tx, err = bases.CreateBases(r, tx, IDTeam)
	if err != nil {
		res = result.SetErrorResult(`Ошибка при создании баз команд`)
		return
	}
	tx, str, err := players.CreatePlayers(r, tx, IDTeam, data.Name, data.Country)
	if err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Ошибка при создании баз команд`)
		return
	}
	query = `INSERT into teams.data (team_id, name, str, players_num, avg_str, cash, cost, price, is_auction) VALUES ($1, $2, $3, 31, $4, 1000000, 50000000, 5, true)`
	params = []interface{}{IDTeam, data.Name, str, float64(str) / 31}
	_, err = tx.ExecContext(ctx, query, params...)
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
	if err = tx.Commit(); err != nil {
		report.ErrorServer(r, err)
		res = result.SetErrorResult(`Внутренняя ошибка`)
		return
	}
	res.Done = true
	res.Items = IDTeam
	return res
}

func CheckCountry(r *http.Request, country string) bool {
	db := config.ConnectDB()
	var exist bool
	query := `SELECT EXISTS(SELECT 1 FROM list.nation_list where name = $1)`
	err := db.QueryRow(query, country).Scan(&exist)
	if err != nil {
		report.ErrorSQLServer(r, err, query, country)
		return false
	}
	return exist
}
