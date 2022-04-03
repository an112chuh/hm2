package get

import (
	"context"
	"hm2/config"
	"hm2/convert"
	"strconv"
)

func TeamsByManager(ctx context.Context, ID int) (TeamID []int, err error) {
	db := config.ConnectDB()
	var TeamNum int
	query := `SELECT team_num FROM list.manager_list WHERE id = $1`
	params := []interface{}{ID}
	err = db.QueryRowContext(ctx, query, params...).Scan(&TeamNum)
	if err != nil {
		return nil, err
	}
	for i := 0; i < TeamNum; i++ {
		var Team int
		query := `SELECT team` + strconv.Itoa(i+1) + ` FROM list.manager_list WHERE id = $1`
		params := []interface{}{ID}
		err = db.QueryRowContext(ctx, query, params...).Scan(&Team)
		if err != nil {
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
		return -1, err
	}
	return TeamID, nil
}
