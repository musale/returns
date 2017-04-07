package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

type InboxData struct {
	From    string
	Code    string
	APIID   string
	Message string
	UserID  string
	APIDate string
}

// InboxPage callback for incoming messages
func InboxPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Println("InboxPage: ", r.Form)

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

	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	jsonReq, err := json.Marshal(request)

	if err != nil {
		log.Println("scheduled to json: ", err)
	}

	if _, err := redisCon.Do("RPUSH", "inbox", string(jsonReq)); err != nil {
		log.Println("inbox queue error: ", err)
	}

	w.WriteHeader(200)
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
				err = saveInbox(&inboxObj)
				if err != nil {
					log.Println("save inbox", err)
				}
			}
		}
	}
	return
}

func saveInbox(req *InboxRequest) error {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	keyName := "code:" + req.From

	codeType, err := redis.String(redisCon.Do("HGET", keyName, "code_type"))

	if err != nil && err == redis.ErrNil {
		// save in hanging messages
		// short code is unassigned
		return err
	}

	if codeType == "DEDICATED" {
		codeUser, err := redis.String(redisCon.Do("HGET", keyName, "user_id"))

		if err != nil && err == redis.ErrNil {
			// save in hanging messages
			return nil
		}
		err = saveMessage(&InboxData{
			From: req.From, Code: req.To, APIID: req.MessageID,
			Message: req.Message, UserID: codeUser, APIDate: req.Date,
		})
		if err != nil {
			log.Println("saveMessage: ", err)
		}
	} else {
		codeKeyword := strings.ToLower(strings.Fields(req.Message)[0])

		codeUser, err := redis.String(redisCon.Do("HGET", keyName, codeKeyword))

		if err != nil && err == redis.ErrNil {
			// save in hanging messages
			return nil
		}
		err = saveMessage(&InboxData{
			From: req.From, Code: req.From, APIID: req.MessageID,
			Message: req.Message, UserID: codeUser, APIDate: req.Date,
		})
		if err != nil {
			log.Println("saveMessage: ", err)
		}
	}
	return nil
}

func saveMessage(req *InboxData) error {

	stmt, err := utils.DBCon.Prepare("insert into bsms_smsinbox(is_read, sender, short_code, api_id, message, user_id, deleted, api_date, insert_date) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Couldn't prepare for inbox insert", err)
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		0, req.From, req.Code, req.APIID, req.Message, req.UserID, 0,
		req.APIDate, time.Now(),
	)

	if err != nil {
		log.Println("Cannot run insert Inbox", err)
		return err
	}

	oid, _ := res.LastInsertId()

	log.Println("Saved Inbox, id:", oid)
	err = sendAutoResponse(req)
	if err != nil {
		return err
	}
	return nil
}

func sendAutoResponse(req) error {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	keyString := "auto:" + req.Keyword, +":" + req.UserID
	autoResp, err := redisCon.Do("GET", keyString)

	if err != nil {
		if err == redis.ErrNil {
			return nil
		} else {
			return err
		}
	}
	var KeyVal map[string]string
	err = json.Unmarshal(keyVal, autoResp)
	if err != nil {
		return err
	}
	log.Println(keyVal)
	userBal, err := getUserBalance(req.UserID)

	if userBal >= 1 {
		// send autoresp sms
		// deduct money
		// save sent sms
		// cache the message for dlr
	} else {
		log.Println("User has no bal for autoresponse")
	}
	return nil
}
