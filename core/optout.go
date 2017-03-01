package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/etowett/returns/common"
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

	common.Logger.Println("Optout request: ", request)

	saveOptout(request)

	fmt.Fprintf(w, "Optout Received")
	return
}

func saveOptout(req map[string]string) {
	stmt, err := common.DbCon.Prepare("insert into callbacks_optout (senderid, phone, time_added) values (?, ?, ?)")
	if err != nil {
		common.Logger.Println("Couldn't prepare for optout insert", err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(req["sid"], req["num"], time.Now())

	if err != nil {
		common.Logger.Println("Cannot run insert optout", err)
		return
	}

	common.Logger.Println("Saved opt out: ", req)

	return
}
