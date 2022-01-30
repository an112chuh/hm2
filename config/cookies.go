package config

import (
	"hm2/constants"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

func InitCookies() {
	constants.SetCookies()
	Store = sessions.NewCookieStore(
		constants.Cookies.AuthKeyOne,
		constants.Cookies.EncryptionOne,
	)
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 30000,
		HttpOnly: true,
	}
}
