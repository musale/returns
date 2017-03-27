package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/etowett/returns/utils"
)

func OptoutPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

	sid := r.FormValue("senderId")
	num := r.FormValue("phoneNumber")

	if len(sid) == 0 || len(num) == 0 {
		fmt.Fprintf(w, "Params not found")
		return
	}

	request := map[string]string{
		"sid": sid, "num": num,
	}

	log.Println("Optout request: ", request)

	saveOptout(request)

	fmt.Fprintf(w, "Optout Received")
	return
}

func saveOptout(req map[string]string) {
	stmt, err := utils.DBCon.Prepare("insert into callbacks_optout (senderid, phone, time_added) values (?, ?, ?)")
	if err != nil {
		log.Println("Couldn't prepare for optout insert", err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(req["sid"], req["num"], time.Now())

	if err != nil {
		log.Println("Cannot run insert optout", err)
		return
	}

	log.Println("Saved opt out: ", req)

	return
}
