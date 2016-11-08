
package main

import (
	"gcllbcks/mylib"
	"log"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"gcllbcks/common"
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

	http.HandleFunc("/dlrs", mylib.DlrPage)
	http.HandleFunc("/inbox", mylib.InboxPage)
	http.HandleFunc("/optout", mylib.OptoutPage)
	log.Fatal(http.ListenAndServe(":5151", nil))

}
