package mylib

import (
	"fmt"
	"log"
	"net/http"

	"gcllbcks/common"
)


func DlrPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

    aid := r.FormValue("id")
    status := r.FormValue("status")
    reason := r.FormValue("reason")

    request := map[string]string {
        "aid": aid, "status": status, "reason": reason,
    }

    go saveDlr(request)

	fmt.Fprintf(w, "Dlr Received")
	return
}

func saveDlr(req map[string]string) {
    db := common.DbCon
	stmt, err1 := db.Prepare("update bsms_smsrecipient set status=?, reason=? where api_id=?")
	if err1 != nil {
		log.Fatal("Couldn't prepare for dlr update", err1)
		return
	}

	defer stmt.Close()

	_, err := stmt.Exec(req["status"], req["reason"], req["aid"])

	if err != nil {
		log.Fatal("Cannot run update dlr", err)
		return
	}
    return
}
