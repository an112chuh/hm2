package inits

import (
	"database/sql"
	"hm2/constants"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB(IsLocal bool) (db *sql.DB) {
	var err error
	if IsLocal {
		constants.SetConnectionConstantsLocal()
		db, err = InitDBLocal(constants.ConnectionLocal)
		if err != nil {
			log.Fatal(`[ERROR] Error in connecting local database: ` + err.Error())
		}
		log.Println(`[SERVER] Connection with local database established`)
	} else {
		constants.SetConnectionConstantsGlobal()
		db, err = InitDBGlobal(constants.ConnectionServer)
		if err != nil {
			log.Fatal(`[ERROR] Error in connecting server database: ` + err.Error())
		}
		log.Println(`[SERVER] Connection with server database established`)
	}
	return db
}

func InitDBLocal(NameDB constants.AuthDB) (db *sql.DB, err error) {
	db, err = sql.Open(NameDB.DBType, NameDB.Login+":"+NameDB.Password+"@/"+NameDB.DBName)
	return db, err
}

func InitDBGlobal(NameDB constants.AuthDB) (db *sql.DB, err error) {
	db, err = sql.Open(NameDB.DBType, NameDB.Login+":"+NameDB.Password+"@"+NameDB.ConnectionExtra+"/"+NameDB.DBName)
	return db, err
}
