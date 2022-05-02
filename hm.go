package main

import (
	"encoding/gob"
	"fmt"
	"hm2/config"
	"hm2/daemon"
	"hm2/result"
	"hm2/routes"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var IsOpeningLocal bool

func main() {
	IsOpeningLocal = false
	var AdminName string
	if len(os.Args) == 2 {
		IsOpeningLocal = true
		AdminName = os.Args[1]
	}
	config.InitRandom()
	config.InitDB(IsOpeningLocal, AdminName)
	config.InitCookies()
	config.InitLoggers()
	go daemon.AuctionWorkerStart()
	//	go daemon.CommandLineStart()
	//	_ = match.TestImportMatch(10, 9)
	//	go daemon.GameDayWorkerStart()
	gob.Register(config.User{})

	routeAll := mux.NewRouter()
	routes.GetAllHandlers(routeAll)
	routeAll.Use(mw)
	routeAll.Methods("OPTIONS").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			res := result.SetErrorResult(`пошёл ты нахуй со своим CORSом`)
			result.ReturnJSON(w, &res)
		})
	http.Handle("/", routeAll)
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
	defer config.Db.Close()
}

func mw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
