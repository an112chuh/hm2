package check

import (
	"context"
	"hm2/config"
	"hm2/report"
)

func ManagerExistByNickName(login string) (bool, error) {
	db := config.ConnectDB()
	query := `SELECT EXISTS(SELECT 1 FROM list.manager_list WHERE login = $1)`
	params := []interface{}{login}
	var exists bool
	err := db.QueryRowContext(context.Background(), query, params...).Scan(&exists)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return false, err
	}
	return exists, nil
}
