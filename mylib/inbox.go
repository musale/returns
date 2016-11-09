package mylib

import (
	"fmt"
	"log"
	"net/http"
	"strings"
    "strconv"
	"time"

	"database/sql"
	"gcllbcks/common"
)

type Code struct {
    Id string
    Type string
    UserId sql.NullInt64
}

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

    fmt.Println("Inbox request: ", request)

    go saveInbox(request)

	fmt.Fprintf(w, "Inbox Received")
	return
}

func saveInbox(req map[string]string) {
    dets := getCodeDets(req["code"])

    if (Code{}) == dets {
        fmt.Println("Inbox no code:", req)
    } else {
        req["code_id"] = dets.Id
        if dets.Type == "DEDICATED" {
            if dets.UserId.Valid {
                req["user_id"] = strconv.Itoa(int(dets.UserId.Int64))
                go saveInboxData(req)
            } else {
                fmt.Println("Dedicated has no user:", req)
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
    // cd := new(Code)
    cd := Code{}
    // err := row.Scan(&cd.Id, &cd.Type, &cd.UserId)
    row.Scan(&cd.Id, &cd.Type, &cd.UserId)
	// if err != nil {
	//    log.Fatal("Couldn't scan select code", err)
	//    return cd
	// }
    return cd
}

func checkShared(req map[string]string) {
    db := common.DbCon
    cd := req["code_id"]
    kw := strings.ToLower(strings.Fields(req["txt"])[0])
    row := db.QueryRow("select user_id from callbacks_shared where code_id=? and keyword=?", cd, kw)

    var uid sql.NullInt64
    // err := row.Scan(&uid)
    row.Scan(&uid)

    // fmt.Println("Uid: ", uid)

	// if err != nil {
    //     log.Fatal("Couldn't scan select shared: ", err)
	//     return
	// }

    if uid.Valid {
        req["user_id"] = strconv.Itoa(int(uid.Int64))
        go saveInboxData(req)
    } else {
        fmt.Println("Shared has no user: ", req)
    }
    return
}

func saveInboxData(req map[string]string) {
    db := common.DbCon
	stmt, err1 := db.Prepare("insert into bsms_smsinbox(is_read, sender, short_code, api_id, message, user_id, deleted, api_date, insert_date) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err1 != nil {
		log.Fatal("Couldn't prepare for inbox insert", err1)
		return
	}

	defer stmt.Close()

	_, err := stmt.Exec(0, req["from"], req["code"], req["aid"], req["txt"], req["user_id"], 0, req["date"], time.Now())

	if err != nil {
		log.Fatal("Cannot run insert Inbox", err)
		return
	}

    fmt.Println("Saved Inbox:", req)

    return
}

