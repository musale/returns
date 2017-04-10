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
	_ = r.ParseForm()
	log.Println("InboxPage: ", r.Form)

	// if r.Method != "POST" {
	// 	w.Header().Set("Allow", "POST")
	// 	w.WriteHeader(http.StatusMethodNotAllowed)
	// 	return
	// }

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

		if err != nil {
			if err == redis.ErrNil {
				time.Sleep(time.Second * 2)
			} else {
				log.Println("ListenForInbox: ", err)
			}
		}

		for _, values := range request {
			if values != "inbox" {
				err := json.Unmarshal([]byte(values), &inboxObj)
				if err != nil {
					log.Println("req Unmarshal", err)
				}
				err = saveInbox(&inboxObj)
				if err != nil {
					log.Println("saveInbox", err)
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

	if err != nil {
		if err == redis.ErrNil {
			// save in hanging messages
			// short code is unassigned
			return nil
		} else {
			return err
		}
	}

	if codeType == "DEDICATED" {
		codeUser, err := redis.String(redisCon.Do("HGET", keyName, "user_id"))

		if err != nil {
			if err == redis.ErrNil {
				// save in hanging messages
				return nil
			} else {
				return err
			}
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

	if err != nil {
		return err
	}

	if userBal >= 1 {
		recsData := utils.PushToAt(
			req.From, keyVal["message"], keyVal["sender_id"])

		var data utils.Costs
		if len(recsData.Recipients) > 0 {
			msgCost, err := getMessageCost(
				req.From, keyVal["message"], req.UserID)
			if err != nil {
				return err
			}
			recCosts := map[string]float64{req.From: msgCost}
			data = utils.GetCosts(recsData.Recipients, recCosts)
		} else {
			data = utils.DummyCosts(req.From)
		}

		costData := data.CostData[0]

		recID, err := saveSentSMS(SMSData{
			SenderID: keyVal["sender_id"], Message: keyVal["message"],
			SendTime: time.Now(), ProcessTime: time.Now(), ReplyCode: "",
			SendType: "AUTO_RESP", UserID: req.UserID, Cost: data.TotalCost,
			Currency: "KES", CostData: costData,
		})
		if err != nil {
			return err
		}
		if len(costData.APIID) > 1 {
			if _, err := redisCon.Do("SETEX", costData.APIID, 1209600000000000, recID); err != nil {
				log.Fatal("cache error ", err)
				return err
			}
		}
	} else {
		log.Println("User has no bal for autoresponse")
	}
	return nil
}

func getUserBalance(userID string) (int, error) {
	var userBal int
	err = utils.DBCon.QueryRow("select sum(trans_amount) from billing_cashtransaction where user_id=?", userID).Scan(&userBal)

	if err != nil {
		log.Println("error query messageid using batch")
		return 0, err
	}

	return userBal, nil
}

func getMessageCost(
	number string, message string, userID string,
) (string, error) {
	pricing := map[string]map[string]float64{
		"0.40": map[string]float64{
			"safaricom": 0.5, "airtel": 0.7, "yu": 1.0, "orange": 1.0,
			"equitel": 1.0, "other": 3.0, "ug": 1.5, "tz": 1.5, "sud": 2.0,
			"zm": 3.0,
		},
		"0.60": map[string]float64{
			"safaricom": 0.6, "airtel": 0.8, "yu": 1.0, "orange": 1.0,
			"equitel": 1.0, "other": 4.0, "ug": 1.6, "tz": 1.6, "sud": 3.0,
			"zm": 3.0,
		},
		"0.80": map[string]float64{
			"safaricom": 0.8, "airtel": 1.0, "yu": 1.0, "orange": 1.0,
			"equitel": 1.0, "other": 4.0, "ug": 1.8, "tz": 1.8, "sud": 3.0,
			"zm": 3.0,
		},
		"1.00": map[string]float64{
			"safaricom": 1.0, "airtel": 1.0, "yu": 1.0, "orange": 1.0,
			"equitel": 1.0, "other": 4.0, "ug": 2.0, "tz": 2.0, "sud": 3.0,
			"zm": 3.0,
		},
		"5.00": map[string]float64{
			"safaricom": 1.5, "airtel": 1.5, "yu": 1.5, "orange": 1.5,
			"equitel": 2.0, "other": 5.0, "ug": 5.0, "tz": 5.0, "sud": 5.0,
			"zm": 3.0,
		},
	}
	userTarrif, err := getUserCost(userID)
	if err != nil {
		return 0.0, err
	}

	net := getNet(number)
	pages := math.Ceil(float64(len(message)) / 160)
	cost := pricing[userTarrif][net] * pages
	return fmt.Sprintf("%.2f", cost), nil
}

func getUserCost(userID string) (string, error) {
	var userPrice float64
	err = utils.DBCon.QueryRow("select user_price from accounts_userprofile where user_id=?", userID).Scan(&userPrice)

	if err != nil {
		log.Println("error query user price")
		return 0.0, err
	}

	return fmt.Sprintf("%.2f", userPrice), nil
}

func getNet(number string) string {
	var net string
	kenNet := map[string]string{
		"1": "safaricom", "2": "safaricom", "3": "airtel", "4": "other",
		"5": "yu", "6": "equitel", "7": "orange", "8": "airtel",
		"9": "safaricom", "0": "safaricom",
	}
	if number[0:3] == "254" {
		net = kenNet[number[4]]
	} else if number[0:3] == "256" {
		net = "ug"
	} else if number[0:3] == "255" {
		net = "tz"
	} else if number[0:3] == "211" {
		net = "sud"
	}
	return net
}

func saveSentSMS(req *SMSData) (string, error) {
	tx, err := utils.DBCon.Begin()
	if err != nil {
		return err
	}
	stmt, err := utils.DBCon.Prepare(
		"insert into bsms_smsoutbox (is_done, is_busy, senderid, content, " +
			"reply_code, time_sent, is_sched, process_time, send_type, " +
			"sender_id, deleted) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		1, 1, req.SenderID, req.Message, req.ReplyCode,
		req.SendTime, 0, req.ProcessTime, req.SendType, req.UserID, 0,
	)
	if err != nil {
		return err
	}

	outboxID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if req.Cost > 0 {
		stmt, err = utils.DBCon.Prepare(
			"insert into billing_cashtransaction (trans_type, ipn_log_id, " +
				"trans_amount, trans_currency, trans_date, user_id) " +
				"values(?, ?, ?, ?, ?, ?)",
		)
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(
			"OUTGOING_MESSAGE", outboxID, req.Cost, req.Currency,
			req.ProcessTime, req.UserID,
		)
		if err != nil {
			return err
		}
	}

	stmt, err = utils.DBCon.Prepare(
		"insert into bsms_smsrecipient (message_content, is_sent, number," +
			"status,reason, api_id, cost, cost_currency, time_sent, " +
			"message_id, user_id, time_processed) values (?, ?, ?, ?, ?," +
			" ?, ?, ?, ?, ?, ?, ?)",
	)

	if err != nil {
		return err
	}
	res, err = stmt.Exec(
		"", 1, req.CostData.Number, rec.CostData.Status,
		rec.CostData.Reason, rec.CostData.APIID, rec.CostData.Cost,
		req.Currency, req.SendTime, outboxID, req.UserID, req.ProcessTime,
	)

	if err != nil {
		return err
	}
	recID, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	tx.Commit()
	return recID, nil
}
