package convert

import (
	"hm2/config"
)

func NationToString(nat int) (string, error) {
	var res string
	db := config.ConnectDB()
	query := `SELECT short from list.nation_list where id = $1`
	params := []interface{}{nat}
	err := db.QueryRow(query, params...).Scan(&res)
	return res, err
}

func NationToLongString(nat int) (string, error) {
	var res string
	db := config.ConnectDB()
	query := `SELECT name from list.nation_list where id = $1`
	params := []interface{}{nat}
	err := db.QueryRow(query, params...).Scan(&res)
	return res, err
}

func PosToString(pos int) string {
	switch pos {
	case 0:
		return "GK"
	case 1:
		return "LD"
	case 2:
		return "RD"
	case 3:
		return "LW"
	case 4:
		return "C"
	case 5:
		return "RW"
	}
	return "ERROR"
}

func NationToInt(nat string) (int, error) {
	var res int
	db := config.ConnectDB()
	query := `SELECT id from list.nation_list where name = $1`
	params := []interface{}{nat}
	err := db.QueryRow(query, params...).Scan(&res)
	return res, err
}
