package config

import (
	"log"
	"strconv"

	"gopkg.in/natefinch/lumberjack.v2"
)

var ErrorLog *log.Logger
var AccessLog *log.Logger

type Rights int

const (
	Default Rights = 1
	Admin   Rights = 2
)

type User struct {
	Username      string
	ID            int `json:"id"`
	Rights        Rights
	Authenticated bool
}

var UserLogs = map[int]*log.Logger{}

func InitLoggers() {
	ErrorFile := &lumberjack.Logger{
		Filename:   "./logs/errors.log",
		MaxSize:    250,
		MaxBackups: 5,
		MaxAge:     10,
	}
	ErrorLog = log.New(ErrorFile, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
	AccessLog = log.New(ErrorFile, "SERVER ", log.Ldate|log.Ltime)
}

func InitUserLogger(id int) {
	UserLogFile := &lumberjack.Logger{
		Filename:   "./logs/users/user_" + strconv.Itoa(id) + ".log",
		MaxSize:    250,
		MaxBackups: 5,
		MaxAge:     10,
	}
	UsersLog := log.New(UserLogFile, "USER: ", log.Ldate|log.Ltime)
	UserLogs[id] = UsersLog
}
