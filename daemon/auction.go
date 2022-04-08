package daemon

import (
	"context"
	"fmt"
	"hm2/config"
	"hm2/get"
	"hm2/report"
	"strconv"
	"time"
)

func AuctionWorkerStart() {
	db := config.ConnectDB()
	query := `SELECT teams.data.team_id FROM teams.data
	INNER JOIN list.team_list ON teams.data.team_id = list.team_list.id 
	WHERE is_auction = true AND manager_id < 0`
	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, nil)
	}
	for rows.Next() {
		var NewTeamID int
		err = rows.Scan(&NewTeamID)
		if err != nil {
			report.ErrorServer(nil, err)
		}
		go AuctionWorkHandler(NewTeamID)
	}
	rows.Close()
	for {
		query := `SELECT teams.data.team_id, teams.data.price FROM teams.data
		INNER JOIN list.team_list ON teams.data.team_id = list.team_list.id 
		WHERE (is_auction = true AND is_started = false) AND manager_id < 0`
		rows, err := db.QueryContext(context.Background(), query)
		if err != nil {
			report.ErrorSQLServer(nil, err, query, nil)
		}
		for rows.Next() {
			var NewTeamID, StartPrice int
			err = rows.Scan(&NewTeamID, &StartPrice)
			if err != nil {
				report.ErrorServer(nil, err)
			}
			query = `UPDATE teams.data SET is_started = true WHERE team_id = $1`
			params := []interface{}{NewTeamID}
			_, err := db.ExecContext(context.Background(), query, params...)
			if err != nil {
				report.ErrorSQLServer(nil, err, query, nil)
			}
			query = `INSERT INTO teams.auction (team_id, manager_id, start_price, bet, put_time, actual) VALUES ($1, -1, $2, -1, $3, true)`
			params = []interface{}{NewTeamID, StartPrice, time.Now()}
			_, err = db.ExecContext(context.Background(), query, params...)
			if err != nil {
				report.ErrorSQLServer(nil, err, query, nil)
			}
			go AuctionWorkHandler(NewTeamID)
		}
		rows.Close()
		time.Sleep(5 * time.Second)
	}
}

func AuctionWorkHandler(TeamID int) {
	for {
		db := config.ConnectDB()
		fmt.Println(TeamID)
		query := `SELECT end_time FROM teams.auction WHERE team_id = $1 and actual = true`
		params := []interface{}{TeamID}
		var EndTime *time.Time
		err := db.QueryRowContext(context.Background(), query, params...).Scan(&EndTime)
		if err != nil {
			report.ErrorSQLServer(nil, err, query, params...)
			time.Sleep(10 * time.Second)
			continue
		}
		if EndTime == nil {
			time.Sleep(10 * time.Second)
			continue
		}
		CurrentTime := time.Now()
		fmt.Println(CurrentTime)
		fmt.Println(EndTime)
		if CurrentTime.After(*EndTime) {
			DiscardChanges(TeamID)
			break
		}
		time.Sleep(10 * time.Second)
	}
}

func DiscardChanges(TeamID int) {
	db := config.ConnectDB()
	fmt.Println("here")
	tx, err := db.Begin()
	if err != nil {
		report.ErrorServer(nil, err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	query := `SELECT manager_id, bet FROM teams.auction_history WHERE team_id = $1 AND actual = true ORDER BY bet DESC OFFSET 1`
	params := []interface{}{TeamID}
	rows, err := db.QueryContext(context.Background(), query, params...)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	defer rows.Close()
	for rows.Next() {
		var ID, bet int
		err = rows.Scan(&ID, &bet)
		if err != nil {
			report.ErrorServer(nil, err)
		}
		query = `UPDATE managers.data SET cash = cash + $1 WHERE id = $2`
		params = []interface{}{bet, ID}
		_, err = tx.ExecContext(context.Background(), query, params...)
		if err != nil {
			report.ErrorSQLServer(nil, err, query, params...)
		}
	}
	var ManagerID int
	query = `SELECT manager_id FROM teams.auction WHERE team_id = $1`
	params = []interface{}{TeamID}
	err = db.QueryRowContext(context.Background(), query, params...).Scan(&ManagerID)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	var TeamNum int
	query = `SELECT team_num FROM list.manager_list WHERE id = $1`
	params = []interface{}{ManagerID}
	err = db.QueryRowContext(context.Background(), query, params...).Scan(&TeamNum)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	query = `UPDATE list.manager_list SET team_num = team_num + 1, team` + strconv.Itoa(TeamNum+1) + ` = $1 WHERE id = $2`
	params = []interface{}{TeamID, ManagerID}
	_, err = tx.ExecContext(context.Background(), query, params...)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	query = `UPDATE list.team_list SET manager_id = $1 WHERE id = $2`
	params = []interface{}{ManagerID, TeamID}
	_, err = tx.ExecContext(context.Background(), query, params...)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	query = `UPDATE teams.auction SET actual = false WHERE team_id = $1`
	params = []interface{}{TeamID}
	_, err = tx.ExecContext(context.Background(), query, params...)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	query = `UPDATE teams.auction_history SET actual = false WHERE team_id = $1`
	_, err = tx.ExecContext(context.Background(), query, params...)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	TeamName, err := get.TeamNameByTeamID(context.Background(), TeamID)
	if err != nil {
		report.ErrorServer(nil, err)
		return
	}
	query = `INSERT into managers.history (manager_id, date_start, team_name, team_id, G, W, WO, WS, LS, LO, L, trophies) VALUES ($1, $2, $3, $4, 0, 0, 0, 0, 0, 0, 0, 0)`
	params = []interface{}{ManagerID, time.Now(), TeamName, TeamID}
	_, err = tx.ExecContext(context.Background(), query, params...)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
	}
	if err = tx.Commit(); err != nil {
		report.ErrorServer(nil, err)
		return
	}
}
