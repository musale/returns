package mylib

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gcllbcks/common"
)


func OptoutPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

    sid := r.FormValue("senderId")
    num := r.FormValue("phoneNumber")

    request := map[string]string {
        "sid": sid, "num": num,
    }

    go saveOptout(request)

	fmt.Fprintf(w, "Optout Received")
	return
}

func saveOptout(req map[string]string) {
    db := common.DbCon
	stmt, err1 := db.Prepare("insert into callbacks_optout (senderid, phone, time_added) values (?, ?, ?)")
	if err1 != nil {
		log.Fatal("Couldn't prepare for optout insert", err1)
		return
	}

	defer stmt.Close()

	_, err := stmt.Exec(req["sid"], req["num"], time.Now())

	if err != nil {
		log.Fatal("Cannot run insert optout", err)
		return
	}
    return
}
