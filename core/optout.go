package core

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/etowett/returns/utils"
)

type OptOutRequest struct {
	SenderID string    `json:"sender_id"`
	Phone    string    `json:"phone_number"`
	Time     time.Time `json:"time"`
}

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

	request := OptOutRequest{
		SenderID: senderID, Phone: phoneNumber, Time: time.Now(),
	}

	log.Println("Optout request: ", request)

	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	jsonReq, err := json.Marshal(request)

	if err != nil {
		log.Println("Optout request: ", err)
	}

	if err := redisCon.Do("RPUSH", "optout", string(jsonReq)); err != nil {
		log.Println("optout queue error: ", err)
	}

	w.WriteHeader(200)
	w.Header().Set("Server", "Returns")
	fmt.Fprintf(w, "Optout Received")
	return
}

// ListenForOptOut on redis
func ListenForOptOut() {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	var optOutReq OptOutRequest

	for {
		request, err := redis.Strings(redisCon.Do("BLPOP", "inbox", 1))

		if err != nil && err == redis.ErrNil {
			time.Sleep(time.Second * 2)
		}

		for _, values := range request {
			if values != "optout" {
				err := json.Unmarshal([]byte(values), &optOutReq)
				if err != nil {
					log.Println("req Unmarshal", err)
				}
				err := saveOptout(&optOutReq)
				if err != nil {
					log.Println("saveOptout", err)
				}
			}
		}
	}
	return
}

func saveOptout(req *OptOutRequest) error {
	stmt, err := utils.DBCon.Prepare("insert into callbacks_optout (senderid, phone, time_added) values (?, ?, ?)")
	if err != nil {
		log.Println("Couldn't prepare for optout insert", err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(req.SenderID, req.Phone, req.Time)

	if err != nil {
		log.Println("Cannot run insert optout", err)
		return err
	}

	log.Println("Saved opt out: ", req)

	return nil
}
