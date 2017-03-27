package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/etowett/returns/utils"
)

func OptoutPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	senderID := r.FormValue("senderId")
	phoneNumber := r.FormValue("phoneNumber")

	if len(senderID) == 0 || len(phoneNumber) == 0 {
		fmt.Fprintf(w, "Required Params not found")
		return
	}

	request := map[string]string{
		"senderID": senderID, "phoneNumber": phoneNumber,
	}

	log.Println("Optout request: ", request)

	saveOptout(request)

	w.WriteHeader(200)
	w.Header().Set("Server", "Returns")
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

	_, err = stmt.Exec(req["senderID"], req["phoneNumber"], time.Now())

	if err != nil {
		log.Println("Cannot run insert optout", err)
		return
	}

	log.Println("Saved opt out: ", req)

	return
}
