package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"hm2/inits"
	"hm2/routes"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type User struct {
	Username      string
	ID            int
	Authenticated bool
}

var store *sessions.CookieStore
var db *sql.DB

func main() {
	IsOpeningLocal := false
	if len(os.Args) == 2 {
		IsOpeningLocal = true
	}
	db = inits.InitDB(IsOpeningLocal)
	store = inits.InitCookies()
	gob.Register(User{})

	routeAll := mux.NewRouter()
	routes.GetAllHandlers(routeAll)
	routeAll.Use(mw)
	var APP_IP, APP_PORT string
	if IsOpeningLocal {
		APP_IP = "127.0.0.1"
		APP_PORT = "8080"
	} else {
		APP_IP = os.Getenv("APP_IP")
		APP_PORT = os.Getenv("APP_PORT")
	}
	fmt.Println("[SERVER] Server address is " + APP_IP + ":" + APP_PORT)
	//	go http.ListenAndServeTLS(APP_IP+":"+APP_PORT, "cert.crt", "key.key", nil)
	http.ListenAndServe(APP_IP+":"+APP_PORT, nil)
	fmt.Println("[SERVER] Server is started")
	defer db.Close()
}

func mw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
