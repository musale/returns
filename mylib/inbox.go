
package mylib

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gcllbcks/common"
)


func InboxPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

    from := r.FormValue("from")
    to := r.FormValue("to")
    text := r.FormValue("text")
    date := r.FormValue("date")
    id := r.FormValue("id")

    request := map[string]string {
        "from": from, "code": to, "txt": text,
        "date": date, "aid": id,
    }

    go saveInbox(request)

	fmt.Fprintf(w, "Inbox Received")
	return
}

type Code struct {
    Id int
    Type string
    UserId string
}

func saveInbox(req map[string]string) {
    dets := getCodeDets(req["code"])

    if dets == (Code{}) {
    } else {
        req["code_id"] = dets.Id
        if dets.Type == "DEDICATED" {
            if dets.UserId != nil {
                req["user_id"] = dets.UserId
                go saveInboxData(req)
            }
        } else if dets.Type == "SHARED" {
            go checkShared(req)
        }
    }
    return
}

func getCodeDets(code string) Code {
    db := common.DbCon
    row := db.QueryRow("select id, code_type, user_id from callbacks_code where code=?", code)
    cd := new (Code)
    err := row.Scan(&cd.Id, &cd.Type, &cd.UserId)
    return cd
}

func checkShared(req map[string]string) {
    db := common.DbCon
    row := db.QueryRow("select user_id from callbacks_shared where code_id=? and keyword=?", req["code_id"], strings.ToLower(req["txt"][0]))

    var uid int
    row.Scan(&uid)

    if uid != nil {
        req["user_id"] = uid
        go saveInboxData(req)
    }
    return
}

func saveInboxData(req map[string]string) {
    db := common.DbCon
	stmt, err1 := db.Prepare("insert into bsms_smsinbox(is_read, sender, short_code, api_id, message, user_id, deleted, api_date, insert_date) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err1 != nil {
		log.Fatal("Couldn't prepare for inbox insert", err)
		return
	}

	defer stmt.Close()

	res, err := stmt.Exec(0, req["from"], req["code"], req["aid"], req["txt"], req["user_id"], 0, req["date"], time.Now())

	if err != nil {
		log.Fatal("Cannot run insert Inbox", err)
		return "Error"
	}

    return
}

