package config

import (
	"hm2/constants"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var Db *sqlx.DB

func InitDB(IsLocal bool) {
	var err error
	if IsLocal {
		constants.SetConnectionConstantsLocal()
		Db, err = InitDBLocal(constants.ConnectionLocal)
		if err != nil {
			log.Fatal(`[ERROR] Error in connecting local database: ` + err.Error())
		}
		log.Println(`[SERVER] Connection with local database established`)
	} else {
		constants.SetConnectionConstantsGlobal()
		Db, err = InitDBGlobal(constants.ConnectionServer)
		if err != nil {
			log.Fatal(`[ERROR] Error in connecting server database: ` + err.Error())
		}
		log.Println(`[SERVER] Connection with server database established`)
	}
}

func InitDBLocal(NameDB constants.AuthDB) (db *sqlx.DB, err error) {
	db, err = sqlx.Open(NameDB.DBType, ConnectionDataLocalCollect())
	return db, err
}

func InitDBGlobal(NameDB constants.AuthDB) (db *sqlx.DB, err error) {
	db, err = sqlx.Open(NameDB.DBType, ConnectionDataGlobalCollect())
	return db, err
}

func ConnectDB() (db *sqlx.DB) {
	return Db
}

func ConnectionDataLocalCollect() (res string) {
	res += "host="
	res += constants.ConnectionLocal.Host
	res += " port="
	res += constants.ConnectionLocal.Port
	res += " user="
	res += constants.ConnectionLocal.Login
	res += " password="
	res += constants.ConnectionLocal.Password
	res += " dbname="
	res += constants.ConnectionLocal.DBName
	res += " connect_timeout=10 sslmode=disable"
	return res
}

func ConnectionDataGlobalCollect() (res string) {
	res = ""
	return res
}
