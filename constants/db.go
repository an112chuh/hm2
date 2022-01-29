package constants

type AuthDB struct {
	Login           string
	Password        string
	DBType          string
	DBName          string
	ConnectionExtra string
}

var ConnectionLocal AuthDB
var ConnectionServer AuthDB

func SetConnectionConstantsLocal() {
	ConnectionLocal.Login = "root"
	ConnectionLocal.Password = "qwerty"
	ConnectionLocal.DBType = "mysql"
}

func SetConnectionConstantsGlobal() {
	//запулить сюда новые константы
}
