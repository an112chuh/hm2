package get

import (
	"context"
	"errors"
	"hm2/config"
	"hm2/convert"
	"hm2/report"
	"strconv"
)

func TeamsByManager(ctx context.Context, ID int) (TeamID []int, err error) {
	db := config.ConnectDB()
	var TeamNum int
	query := `SELECT team_num FROM list.manager_list WHERE id = $1`
	params := []interface{}{ID}
	err = db.QueryRowContext(ctx, query, params...).Scan(&TeamNum)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return nil, err
	}
	for i := 0; i < TeamNum; i++ {
		var Team int
		query := `SELECT team` + strconv.Itoa(i+1) + ` FROM list.manager_list WHERE id = $1`
		params := []interface{}{ID}
		err = db.QueryRowContext(ctx, query, params...).Scan(&Team)
		if err != nil {
			report.ErrorSQLServer(nil, err, query, params...)
			return nil, err
		}
		TeamID = append(TeamID, Team)
	}
	return TeamID, nil
}

func NationIDByTeam(ctx context.Context, TeamID int) (int, error) {
	db := config.ConnectDB()
	var TeamString string
	var NationID int
	query := `SELECT country from list.team_list where id = $1`
	params := []interface{}{TeamID}
	err := db.QueryRowContext(ctx, query, params...).Scan(&TeamString)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return -1, err
	}
	NationID, err = convert.NationToInt(TeamString)
	if err != nil {
		return -1, err
	}
	return NationID, nil
}

func TeamIDByTeamName(ctx context.Context, TeamName string) (int, error) {
	db := config.ConnectDB()
	var TeamID int
	query := `SELECT id FROM list.team_list where name = $1`
	params := []interface{}{TeamName}
	err := db.QueryRowContext(ctx, query, params...).Scan(&TeamID)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return -1, err
	}
	return TeamID, nil
}

func TeamNameByTeamID(ctx context.Context, TeamID int) (string, error) {
	db := config.ConnectDB()
	var TeamName string
	query := `SELECT name FROM list.team_list where id = $1`
	params := []interface{}{TeamID}
	err := db.QueryRowContext(ctx, query, params...).Scan(&TeamName)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return ``, err
	}
	return TeamName, nil
}

func CurrentTeamByManager(ctx context.Context, IDManager int) (int, error) {
	db := config.ConnectDB()
	var CurTeamNum int
	query := `SELECT cur_team FROM list.manager_list WHERE id = $1`
	params := []interface{}{IDManager}
	err := db.QueryRowContext(ctx, query, params...).Scan(&CurTeamNum)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return -1, err
	}
	if CurTeamNum < 1 || CurTeamNum > 3 {
		err = errors.New(`номер текущей команды должен быть от 1 до 3`)
		report.ErrorServer(nil, err)
		return -1, err
	}
	CurTeamNumString := strconv.Itoa(CurTeamNum)
	var TeamID int
	query = `SELECT team` + CurTeamNumString + ` FROM list.manager_list WHERE id = $1`
	params = []interface{}{IDManager}
	err = db.QueryRowContext(ctx, query, params...).Scan(&TeamID)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return -1, err
	}
	if TeamID < 1 {
		err = errors.New(`У менеджера меньше команд, чем указано`)
		report.ErrorServer(nil, err)
		return -1, err
	}
	return TeamID, nil
}
