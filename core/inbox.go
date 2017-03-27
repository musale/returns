package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "strconv"
	// "strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/etowett/returns/utils"
)

type Code struct {
	Id     string
	Type   string
	UserId sql.NullInt64
}

type InboxRequest struct {
	From string
	To string
	Message string
	Date string
	MessageID string
}

func InboxPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	from := r.FormValue("from")
	to := r.FormValue("to")
	text := r.FormValue("text")
	date := r.FormValue("date")
	id := r.FormValue("id")

	// request := map[string]string {
	// 	"from": from, "code": to, "txt": text,
	// 	"date": date, "aid": id,
	// }

	request := InboxRequest {
		From: from, To: to, Message: text, Date: date, MessageID: id,
	}

	log.Println("Inbox request: ", request)

	queueInbox(request)

	w.WriteHeader(200)
	w.Header().Set("Server", "Returns")
	fmt.Fprintf(w, "RMDlr Received")
	fmt.Fprintf(w, "Inbox Received")
	return
}

func queueInbox(request InboxRequest) {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	jsonReq, err := json.Marshal(request)

	if err != nil {
		log.Println("scheduled to json: ", err)
	}

	redisCon.Do("RPUSH", "inbox", string(jsonReq))
	return
}

// ListenForInbox on redis
func ListenForInbox() {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	var inboxObj InboxRequest

	for {
		request, err := redis.Strings(redisCon.Do("BLPOP", "inbox", 1))

		if err != nil && err == redis.ErrNil {
			time.Sleep(time.Second * 2)
		}

		for _, values := range request {
			if values != "dlrs" {
				err := json.Unmarshal([]byte(values), &inboxObj)
				if err != nil {
					log.Println("req Unmarshal", err)
				}
				// saveInbox(&inboxObj)
				log.Println("Inbox: ", inboxObj)
			}
		}
	}
}

// func saveInbox(req *InboxRequest) {

// 	// cache codes as redis hashes code:31390 type shared kip 1 vic 3 steph 7171

// 	dets := getCodeDets(req["code"])

// 	if (Code{}) == dets {
// 		log.Println("Inbox no code:", req)
// 	} else {
// 		req["code_id"] = dets.Id
// 		if dets.Type == "DEDICATED" {
// 			if dets.UserId.Valid {
// 				req["user_id"] = strconv.Itoa(int(dets.UserId.Int64))
// 				go saveInboxData(req)
// 			} else {
// 				log.Println("Dedicated has no user:", req)
// 			}
// 		} else if dets.Type == "SHARED" {
// 			go checkShared(req)
// 		}
// 	}
// 	return
// }

// func getCodeDets(code string) Code {
// 	db := utils.DBCon
// 	row := db.QueryRow("select id, code_type, user_id from callbacks_code where code=?", code)
// 	// cd := new(Code)
// 	cd := Code{}
// 	// err := row.Scan(&cd.Id, &cd.Type, &cd.UserId)
// 	row.Scan(&cd.Id, &cd.Type, &cd.UserId)
// 	// if err != nil {
// 	//    logger.Println("Couldn't scan select code", err)
// 	//    return cd
// 	// }
// 	return cd
// }

// func checkShared(req map[string]string) {

// 	db := utils.DBCon
// 	cd := req["code_id"]
// 	kw := strings.ToLower(strings.Fields(req["txt"])[0])
// 	row := db.QueryRow("select user_id from callbacks_shared where code_id=? and keyword=?", cd, kw)

// 	var uid sql.NullInt64
// 	// err := row.Scan(&uid)
// 	row.Scan(&uid)

// 	if uid.Valid {
// 		req["user_id"] = strconv.Itoa(int(uid.Int64))
// 		req["kw"] = kw
// 		go saveInboxData(req)
// 	} else {
// 		log.Println("Shared has no user: ", req)
// 	}
// 	return
// }

// func saveInboxData(req map[string]string) {

// 	db := utils.DBCon
// 	stmt, err1 := db.Prepare("insert into bsms_smsinbox(is_read, sender, short_code, api_id, message, user_id, deleted, api_date, insert_date) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
// 	if err1 != nil {
// 		log.Println("Couldn't prepare for inbox insert", err1)
// 		return
// 	}

// 	defer stmt.Close()

// 	res, err := stmt.Exec(0, req["from"], req["code"], req["aid"], req["txt"], req["user_id"], 0, req["date"], time.Now())

// 	if err != nil {
// 		log.Println("Cannot run insert Inbox", err)
// 		return
// 	}

// 	oid, _ := res.LastInsertId()

// 	log.Println("Saved Inbox, id:", oid)
// 	go sendAutoResponse(req)
// 	return
// }

// func sendAutoResponse(req map[string]string) {
// 	// select * from callbacks_autoresponse where user_id=req['user_id']
// 	// and key=req['kw']

// 	// select sum(trans_amount) from billing_cashtransaction where user_id=uid
// 	// get cost of message
// 	// if bal >= cost
// 	// push to api
// 	// create cashtrans
// 	// save outbox and recipient
// 	return
// }
