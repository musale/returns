package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/etowett/returns/common"
	"github.com/etowett/returns/mylib"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var err error

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	common.DbCon, err = sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASS")+"@tcp("+os.Getenv("DB_HOST")+":3306)/"+os.Getenv("DB_NAME")+"?charset=utf8")
	if err != nil {
		panic(err.Error())
	}
	defer common.DbCon.Close()

	// Test the connection to the database
	err = common.DbCon.Ping()
	if err != nil {
		panic(err.Error())
	}

	logFile, openErr1 := os.OpenFile(os.Getenv("LOG_DIR"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)

	if openErr1 != nil {
		log.Println("Uh oh! Could not open log file.", openErr1)
	}

	defer logFile.Close()

	common.Logger = log.New(logFile, "", log.Lshortfile|log.Ldate|log.Ltime)

	// Listen for Dlrs
	go mylib.ListenForDlrs()

	// Route set up
	http.HandleFunc("/at-dlrs", mylib.ATDlrPage)
	http.HandleFunc("/rm-dlrs", mylib.RMDlrPage)
	http.HandleFunc("/cache-dlr", mylib.CacheDlrPage)
	http.HandleFunc("/inbox", mylib.InboxPage)
	http.HandleFunc("/optout", mylib.OptoutPage)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}
