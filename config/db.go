package config

import (
	"hm2/constants"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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
	db, err = sqlx.Open(NameDB.DBType, NameDB.Login+":"+NameDB.Password+"@/"+NameDB.DBName)
	return db, err
}

func InitDBGlobal(NameDB constants.AuthDB) (db *sqlx.DB, err error) {
	db, err = sqlx.Open(NameDB.DBType, NameDB.Login+":"+NameDB.Password+"@"+NameDB.ConnectionExtra+"/"+NameDB.DBName)
	return db, err
}
