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
	"math/rand"
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
	res = CreateTeamConfirm(r, data, false)
	result.ReturnJSON(w, &res)
}

func CreateTeamConfirm(r *http.Request, data CreateTeamData, IsDebug bool) (res result.ResultInfo) {
	db := config.ConnectDB()
	tx, err := db.Begin()
	if err != nil {
		report.ErrorServer(r, err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	var ctx context.Context
	if r != nil {
		ctx = r.Context()
	} else {
		ctx = context.Background()
	}
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
	var cash, cost, price int
	isauc := true
	if !IsDebug {
		cash = 1000000
		cost = 50000000
		price = 5
	} else {
		str = 1900 + rand.Intn(500)
		cash = 800000 + rand.Intn(400000)
		cost = 40000000 + rand.Intn(20000000)
		price = rand.Intn(20)
		tmp := rand.Intn(3)
		if tmp == 0 || price == 0 {
			isauc = false
		}
	}
	query = `INSERT into teams.data (team_id, name, str, players_num, avg_str, cash, cost, price, is_auction) VALUES ($1, $2, $3, 31, $4, $5, $6, $7, $8)`
	params = []interface{}{IDTeam, data.Name, str, float64(str) / 31, cash, cost, price, isauc}
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
