package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/etowett/returns/core"
	"github.com/etowett/returns/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file ", err)
	}

	f, err := os.OpenFile(os.Getenv("LOG_FILE"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("Log file error: ", err)
	}
	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//set output of logs to f
	log.SetOutput(f)

	utils.DBCon, err = sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASS")+"@tcp("+os.Getenv("DB_HOST")+":3306)/"+os.Getenv("DB_NAME")+"?charset=utf8")
	if err != nil {
		log.Fatal("db error: ", err)
	}
	defer utils.DBCon.Close()

	// Test the connection to the database
	err = utils.DBCon.Ping()
	if err != nil {
		log.Fatal("Error DB ping ", err)
	}

	// Common redis connection object
	utils.RedisCon = utils.RedisPool().Get()
	defer utils.RedisCon.Close()

	// Listen for Dlrs
	go core.ListenForDlrs()
	// Listen for Inbox
	go core.ListenForInbox()
	// Push Scheduled Dlrs to reqdy queue
	go core.PushToQueue()

	// Route set up
	http.HandleFunc("/at-dlrs", core.ATDlrPage)
	http.HandleFunc("/rm-dlrs", core.RMDlrPage)
	http.HandleFunc("/cache-dlr", core.CacheDlrPage)
	http.HandleFunc("/inbox", core.InboxPage)
	http.HandleFunc("/optout", core.OptoutPage)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
