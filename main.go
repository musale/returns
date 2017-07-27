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
	defer f.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	initStuff()

	// Route set up
	http.HandleFunc("/at-dlrs", core.ATDlrPage)
	http.HandleFunc("/rm-dlrs", core.RMDlrPage)
	http.HandleFunc("/saf-dlrs", core.SafDlrPage)
	http.HandleFunc("/saf-notify", core.SafNotifyPage)
	http.HandleFunc("/cache-dlr", core.CacheDlrPage)
	http.HandleFunc("/cache-bulk-dlr", core.CacheBulkDlrPage)
	http.HandleFunc("/inbox", core.InboxPage)
	http.HandleFunc("/optout", core.OptoutPage)
	http.HandleFunc("/saf-inbox", core.SafInboxPage)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func initStuff() {
	log.Println("Started callbacks on port", os.Getenv("PORT"))
	// Listen for Dlrs
	go core.ListenForDlrs()
	// Listen for Inbox
	go core.ListenForInbox()
	// Listen for optout
	go core.ListenForOptOut()
	// Push Scheduled Dlrs to reqdy queue
	go core.PushToQueue()

	startQueueDlrWorkers()
	startSMSPool()
}

func startSMSPool() {
	for i := 1; i <= 10; i++ {
		go func() {
			for notif := range core.SMSNotifs {
				err := notif.ProcessSMSNofit()
				if err != nil {
					log.Println("ProcessSMSNofit: ", err)
				}
			}
		}()
	}
}

func startQueueDlrWorkers() {
	for i := 1; i <= 9; i++ {
		go func() {
			for dlr := range core.DLRReqChan {
				err := core.QueueDlr(dlr)
				if err != nil {
					log.Println("QueueDlr: ", err)
				}
			}
		}()
	}
}
