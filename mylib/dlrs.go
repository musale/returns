package mylib

import (
	"fmt"
	"net/http"

	"github.com/etowett/returns/common"
)

func DlrPage(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger
	if r.Method != "POST" {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

	aid := r.FormValue("id")
	status := r.FormValue("status")
	reason := r.FormValue("reason")

	request := map[string]string{
		"aid": aid, "status": status, "reason": reason,
	}

	go saveDlr(request)

	logger.Println("Dlr Request: ", request)

	fmt.Fprintf(w, "Dlr Received")
	return
}

func saveDlr(req map[string]string) {
	logger := common.Logger
	db := common.DbCon
	stmt, err1 := db.Prepare("update bsms_smsrecipient set status=?, reason=? where api_id=?")
	if err1 != nil {
		logger.Fatal("Couldn't prepare for dlr update", err1)
		return
	}

	defer stmt.Close()

	_, err := stmt.Exec(req["status"], req["reason"], req["aid"])

	if err != nil {
		logger.Println("Cannot run update dlr", err)
		return
	}
	logger.Println("Dlr saved: ", req["aid"])
	return
}
