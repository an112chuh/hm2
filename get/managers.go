package get

import (
	"context"
	"hm2/config"
	"hm2/report"
)

func ManagerByID(ID int) (string, error) {
	db := config.ConnectDB()
	var login string
	query := `SELECT login FROM list.manager_list WHERE id = $1`
	params := []interface{}{ID}
	err := db.QueryRowContext(context.Background(), query, params...).Scan(&login)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return ``, err
	}
	return login, err
}

func ManagerByLogin(Login string) (int, error) {
	db := config.ConnectDB()
	var ID int
	query := `SELECT id FROM list.manager_list WHERE login = $1 AND is_active = TRUE`
	params := []interface{}{Login}
	err := db.QueryRowContext(context.Background(), query, params...).Scan(&ID)
	if err != nil {
		report.ErrorSQLServer(nil, err, query, params...)
		return -1, err
	}
	return ID, err
}
