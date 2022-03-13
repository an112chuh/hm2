package teams

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
	"time"

	"github.com/gorilla/mux"
)

type ManagerTeams struct {
	Team1 int `db:"team1"`
	Team2 int `db:"team2"`
	Team3 int `db:"team3"`
}

func BuyTeamHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, true)
	if !user.Authenticated {
		res = result.SetErrorResult(`Вы не можете купить команду. Пожалуйста, войдите или создайте аккаунт`)
		result.ReturnJSON(w, &res)
		return
	}
	vars := mux.Vars(r)
	idString := vars[`id`]
	ID, _ := strconv.Atoi(idString)
	res = BuyTeam(r, ID, user)
	result.ReturnJSON(w, &res)
}

func SellTeamHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	user := managers.IsLogin(w, r, true)
	if !user.Authenticated {
		res = result.SetErrorResult(`Вы не можете отказаться от команды. Пожалуйста, войдите или создайте аккаунт`)
		result.ReturnJSON(w, &res)
		return
	}
	vars := mux.Vars(r)
	idString := vars[`id`]
	ID, _ := strconv.Atoi(idString)
	res = SellTeam(r, ID, user)
	result.ReturnJSON(w, &res)
}

func AucTeamHandler(w http.ResponseWriter, r *http.Request) {

}

func EditTeamsHandler(w http.ResponseWriter, r *http.Request) {

}

func EditTeamsConfirmHandler(w http.ResponseWriter, r *http.Request) {

}

func BuyTeam(r *http.Request, IDTeam int, user config.User) (res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
	var ManagerCoins, TeamPrice int
	query := `SELECT cash from managers.data where id = $1`
	params := []interface{}{user.ID}
	err := db.QueryRowContext(ctx, query, params...).Scan(&ManagerCoins)
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
	var IsAuction bool
	var NewTeamName string
	query = `SELECT name, price, is_auction from teams.data where team_id = $1`
	params = []interface{}{IDTeam}
	err = db.QueryRowContext(ctx, query, params...).Scan(&NewTeamName, &TeamPrice, &IsAuction)
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
	if IsAuction {
		res = result.SetErrorResult("Невозможно купить аукционную команду")
		return
	}
	if ManagerCoins < TeamPrice {
		res = result.SetErrorResult("Не хватает средств для покупки команды")
		return
	}
	var VipLvl, TeamNum int
	var TeamsHave ManagerTeams
	query = `SELECT vip, team_num, team1, team2, team3 from list.manager_list where id = $1`
	params = []interface{}{user.ID}
	err = db.QueryRowContext(ctx, query, params...).Scan(&VipLvl, &TeamNum, &TeamsHave.Team1, &TeamsHave.Team2, &TeamsHave.Team3)
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
	var AvaliableTeams int
	query = `SELECT teams from managers.vip_levels where vip_lvl = $1`
	params = []interface{}{VipLvl}
	err = db.QueryRowContext(ctx, query, params...).Scan(&AvaliableTeams)
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
	if AvaliableTeams <= TeamNum {
		if VipLvl < 3 {
			res = result.SetErrorResult("Достигнут лимит для взятия команд. Приобретите VIP более высокого уровня, чтобы взять больше команд")
			return
		}
		res = result.SetErrorResult("Достигнут лимит для взятия команд. Пожалуйста, откажитесь от одной из команд перед покупкой другой")
		return
	}
	NewTeamNation, err := GetNationIDByTeam(r, ctx, IDTeam)
	if err != nil {
		res = result.SetErrorResult(err.Error())
		return
	}
	var TeamNationsExist []int
	if TeamsHave.Team1 > 0 {
		HaveTeamNat, err := GetNationIDByTeam(r, ctx, TeamsHave.Team1)
		if err != nil {
			res = result.SetErrorResult(err.Error())
			return
		}
		TeamNationsExist = append(TeamNationsExist, HaveTeamNat)
	}
	if TeamsHave.Team2 > 0 {
		HaveTeamNat, err := GetNationIDByTeam(r, ctx, TeamsHave.Team2)
		if err != nil {
			res = result.SetErrorResult(err.Error())
			return
		}
		TeamNationsExist = append(TeamNationsExist, HaveTeamNat)
	}
	if TeamsHave.Team3 > 0 {
		HaveTeamNat, err := GetNationIDByTeam(r, ctx, TeamsHave.Team3)
		if err != nil {
			res = result.SetErrorResult(err.Error())
			return
		}
		TeamNationsExist = append(TeamNationsExist, HaveTeamNat)
	}
	for i := 0; i < len(TeamNationsExist); i++ {
		if NewTeamNation == TeamNationsExist[i] {
			res = result.SetErrorResult("Нельзя брать команды той же нации, что и предыдущие")
			return
		}
	}
	var ManagerID int
	query = `SELECT manager_id from list.team_list where id = $1`
	params = []interface{}{IDTeam}
	err = db.QueryRowContext(ctx, query, params...).Scan(&ManagerID)
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
	if ManagerID > 0 {
		res = result.SetErrorResult("Команда принадлежит другому менеджеру")
		return
	}
	tx, err := db.Begin()
	if err != nil {
		report.ErrorServer(r, err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	query = `UPDATE managers.data SET cash = cash - $1 WHERE id = $2`
	params = []interface{}{TeamPrice, user.ID}
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
	TeamNumString := strconv.Itoa(len(TeamNationsExist) + 1)
	query = `UPDATE list.manager_list SET team` + TeamNumString + ` = $1, team_num = team_num + 1 WHERE id = $2`
	params = []interface{}{IDTeam, user.ID}
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
	query = `INSERT into managers.history (manager_id, date_start, team_name, team_id, G, W, WO, WS, LS, LO, L, trophies) VALUES ($1, $2, $3, $4, 0, 0, 0, 0, 0, 0, 0, 0)`
	params = []interface{}{user.ID, time.Now(), NewTeamName, IDTeam}
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
	query = `UPDATE list.team_list SET manager_id = $1 WHERE id = $2`
	params = []interface{}{user.ID, IDTeam}
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
	return
}

func SellTeam(r *http.Request, IDTeam int, user config.User) (res result.ResultInfo) {
	db := config.ConnectDB()
	ctx := r.Context()
	var ManagerID int
	query := `SELECT manager_id from list.team_list where id = $1`
	params := []interface{}{IDTeam}
	err := db.QueryRowContext(ctx, query, params...).Scan(&ManagerID)
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
	if ManagerID != user.ID {
		res = result.SetErrorResult("Нельзя отказаться от управления не своей командой")
		return
	}
	var Price int
	query = `SELECT price from teams.data where team_id = $1`
	params = []interface{}{IDTeam}
	err = db.QueryRowContext(ctx, query, params...).Scan(&Price)
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
	var mt ManagerTeams
	query = `SELECT team1, team2, team3 from list.manager_list where id = $1`
	params = []interface{}{user.ID}
	err = db.QueryRowxContext(ctx, query, params...).StructScan(&mt)
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
	var queryString string
	if mt.Team1 == IDTeam {
		queryString = `team1 = team2, team2 = team3, team3 = 0`
	}
	if mt.Team2 == IDTeam {
		queryString = `team2 = team3, team3 = 0`
	}
	if mt.Team3 == IDTeam {
		queryString = `team3 = 0`
	}
	tx, err := db.Begin()
	if err != nil {
		report.ErrorServer(r, err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	query = `UPDATE list.manager_list SET ` + queryString + `, team_num = team_num - 1 where id = $1`
	params = []interface{}{user.ID}
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
	ReturnedPrice := float64(Price) / 2
	query = `UPDATE managers.data SET cash = cash + $1 WHERE id = $2`
	params = []interface{}{ReturnedPrice, user.ID}
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
	var IDRecord int
	query = `SELECT id from managers.history WHERE manager_id = $1 AND team_id = $2 ORDER BY id DESC LIMIT 1`
	params = []interface{}{user.ID, IDTeam}
	err = tx.QueryRowContext(ctx, query, params...).Scan(&IDRecord)
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
	query = `UPDATE managers.history SET date_finish = $1 WHERE manager_id = $2 AND team_id = $3 AND id = $4`
	params = []interface{}{time.Now(), user.ID, IDTeam, IDRecord}
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
	query = `UPDATE list.team_list SET manager_id = -1 WHERE id = $1`
	params = []interface{}{IDTeam}
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
	return
}

func GetNationIDByTeam(r *http.Request, ctx context.Context, TeamID int) (int, error) {
	db := config.ConnectDB()
	var TeamString string
	var NationID int
	query := `SELECT country from list.team_list where id = $1`
	params := []interface{}{TeamID}
	err := db.QueryRowContext(ctx, query, params...).Scan(&TeamString)
	if err != nil {
		switch {
		case errors.Is(ctx.Err(), context.Canceled), errors.Is(ctx.Err(), context.DeadlineExceeded):
			return -1, errors.New(report.CtxError)
		default:
			report.ErrorSQLServer(r, err, query, params...)
			return -1, errors.New(report.UnknownError)
		}
	}
	NationID, err = convert.NationToInt(TeamString)
	if err != nil {
		report.ErrorServer(r, err)
		return -1, err
	}
	return NationID, nil
}
