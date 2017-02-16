package mylib

import (
	"fmt"
	"net/http"
	"time"

	"github.com/etowett/returns/common"
)

func OptoutPage(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger
	if r.Method != "POST" {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

	sid := r.FormValue("senderId")
	num := r.FormValue("phoneNumber")

	request := map[string]string{
		"sid": sid, "num": num,
	}

	logger.Println("Optout request: ", request)

	go saveOptout(request)

	fmt.Fprintf(w, "Optout Received")
	return
}

func saveOptout(req map[string]string) {
	logger := common.Logger
	db := common.DbCon
	stmt, err1 := db.Prepare("insert into callbacks_optout (senderid, phone, time_added) values (?, ?, ?)")
	if err1 != nil {
		logger.Println("Couldn't prepare for optout insert", err1)
		return
	}

	defer stmt.Close()

	_, err := stmt.Exec(req["sid"], req["num"], time.Now())

	if err != nil {
		logger.Println("Cannot run insert optout", err)
		return
	}

	logger.Println("Saved opt out: ", req)

	return
}
