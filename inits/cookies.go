package inits

import (
	"hm2/constants"

	"github.com/gorilla/sessions"
)

func InitCookies() (store *sessions.CookieStore) {
	constants.SetCookies()
	store = sessions.NewCookieStore(
		constants.Cookies.AuthKeyOne,
		constants.Cookies.EncryptionOne,
	)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 30000,
		HttpOnly: true,
	}
	return store
}
