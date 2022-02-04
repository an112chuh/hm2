package constants

type AuthDB struct {
	Login           string
	Password        string
	DBType          string
	Port            string
	Host            string
	DBName          string
	ConnectionExtra string
}

var ConnectionLocal AuthDB
var ConnectionServer AuthDB

func SetConnectionConstantsLocal(AdminName string) {
	ConnectionLocal.Login = "postgres"
	ConnectionLocal.Password = "derwes"
	ConnectionLocal.DBType = "postgres"
	ConnectionLocal.Host = "localhost"
	ConnectionLocal.Port = "5432"
	if AdminName == "st411ar" {
		ConnectionLocal.DBName = "postgres"
	} else {
		ConnectionLocal.DBName = "hm"
	}
}

func SetConnectionConstantsGlobal() {
	//запулить сюда новые константы
}
