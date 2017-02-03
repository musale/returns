package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"returns/common"
	"returns/mylib"

	_ "github.com/go-sql-driver/mysql"
)

var err error

func main() {

	common.DbCon, err = sql.Open("mysql", "kip:kip@db@/smsleopard")
	if err != nil {
		panic(err.Error())
	}
	defer common.DbCon.Close()

	// Test the connection to the database
	err = common.DbCon.Ping()
	if err != nil {
		panic(err.Error())
	}

	logFile, openErr1 := os.OpenFile("logs/callbacks.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)

	if openErr1 != nil {
		log.Println("Uh oh! Could not open log file.", openErr1)
	}

	defer logFile.Close()

	common.Logger = log.New(logFile, "", log.Lshortfile|log.Ldate|log.Ltime)

	// Route set up
	http.HandleFunc("/at-dlrs", mylib.DlrPage)
	http.HandleFunc("/inbox", mylib.InboxPage)
	http.HandleFunc("/optout", mylib.OptoutPage)
	log.Fatal(http.ListenAndServe(":4147", nil))
}
