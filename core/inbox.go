package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/etowett/returns/utils"
	"github.com/garyburd/redigo/redis"
)

type Code struct {
	Id     string
	Type   string
	UserId sql.NullInt64
}

type InboxRequest struct {
	From      string
	To        string
	Message   string
	Date      string
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

	request := InboxRequest{
		From: from, To: to, Message: text, Date: date, MessageID: id,
	}

	log.Println("Inbox request: ", request)

	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	jsonReq, err := json.Marshal(request)

	if err != nil {
		log.Println("scheduled to json: ", err)
	}

	if err := redisCon.Do("RPUSH", "inbox", string(jsonReq)); err != nil {
		log.Println("inbox queue error: ", err)
	}

	w.WriteHeader(200)
	w.Header().Set("Server", "Returns")
	fmt.Fprintf(w, "Inbox Received")
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
			if values != "inbox" {
				err := json.Unmarshal([]byte(values), &inboxObj)
				if err != nil {
					log.Println("req Unmarshal", err)
				}
				saveInbox(&inboxObj)
			}
		}
	}
}

func saveInbox(req *InboxRequest) {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	keyName := "code:" + req.Code

	codeType, err := redis.String(redisCon.Do("HGET", keyName, "code_type"))

	if err != nil && err == redis.ErrNil {
		// save in hanging messages
		return
	}

	if codeType == "DEDICATED" {
		codeUser, err := redis.String(redisCon.Do("HGET", keyName, "user_id"))

		if err != nil && err == redis.ErrNil {
			// save in hanging messages
			return
		}
		saveMessage(&InboxData{
			From: req.From, Code: req.Code, APIID: req.APIID,
			Message: req.Message, UserID: codeUser, APIDate: req.Date,
		})
	} else {
		codeKeyword := strings.ToLower(strings.Fields(req.Message)[0])

		codeUser, err := redis.String(redisCon.Do("HGET", keyName, codeKeyword))

		if err != nil && err == redis.ErrNil {
			// save in hanging messages
			return
		}
		saveMessage(&InboxData{
			From: req.From, Code: req.Code, APIID: req.APIID,
			Message: req.Message, UserID: codeUser, APIDate: req.Date,
		})
	}

}

func saveInboxData(req *InboxData) {

	db := utils.DBCon
	stmt, err1 := db.Prepare("insert into bsms_smsinbox(is_read, sender, short_code, api_id, message, user_id, deleted, api_date, insert_date) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err1 != nil {
		log.Println("Couldn't prepare for inbox insert", err1)
		return
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		0, reqFrom, req.Code, req.APIID, req.Message, req.UserID, 0,
		req.APIDate, time.Now(),
	)

	if err != nil {
		log.Println("Cannot run insert Inbox", err)
		return
	}

	oid, _ := res.LastInsertId()

	log.Println("Saved Inbox, id:", oid)
	go sendAutoResponse(req)
	return
}
