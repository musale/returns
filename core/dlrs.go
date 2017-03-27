package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/etowett/returns/utils"
	"github.com/garyburd/redigo/redis"
)

// DlrRequest struct
type DlrRequest struct {
	APIID, Status, Reason string
	TimeReceived          time.Time
	Retries               int64
}

// DlrRequestInterface definition
type DlrRequestInterface interface {
	parseRequestString() string
	parseRequestMap() map[string]string
}

// Function to parse a DlrRequest item to string
func (request *DlrRequest) parseRequestString() string {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		log.Fatal("Dlr request to json string error", err)
	}
	return string(requestJSON)
}

// Function to parse a DlrRequest item to map[string]string
func (request *DlrRequest) parseRequestMap() map[string]string {
	return map[string]string{
		"api_id": request.APIID,
		"status": request.Status,
		"reason": request.Reason,
	}
}

// ATDlrPage rendering
func ATDlrPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != utils.POST {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Todo: print all params received 

	apiID := r.FormValue("id")
	apiStatus := r.FormValue("status")

	if strings.ToUpper(apiStatus) == "SUCCESS" {
		apiStatus = "DELIVRD"
	}

	request := DlrRequest{
		APIID: apiID, Status: strings.ToUpper(apiStatus),
		TimeReceived: time.Now(), Retries: 0,
	}

	if apiStatus == "Failed" || apiStatus == "Rejected" {
		request.Reason = r.FormValue("failureReason")
	}

	log.Println("ATDLR Request:", request)

	go pushToQueue(&request)

	w.WriteHeader(200)
	w.Header().Set("Server", "Returns")
	fmt.Fprintf(w, "ATDlr Received")
	return
}

// RMDlrPage rendering
func RMDlrPage(w http.ResponseWriter, r *http.Request) {

	if r.Method != utils.POST {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	for key, values := range r.Form {   // range over map
		for _, value := range values {    // range over []string
			log.Println(key, value)
		}
	}

	apiID := r.FormValue("sMessageId")
	apiStatus := r.FormValue("sStatus")
	// senderID := r.FormValue("sSender")
	// phoneNumber := r.FormValue("sMobileNo")
	// dateDone := r.FormValue("dtDone")
	// dateSubmitted := r.FormValue("dtSubmit")

	request := DlrRequest{
		APIID: apiID, Status: strings.ToUpper(apiStatus),
		TimeReceived: time.Now(), Retries: 0,
	}

	log.Println("RMDLR Request:", request)

	go pushToQueue(&request)

	w.WriteHeader(200)
	w.Header().Set("Server", "Returns")
	fmt.Fprintf(w, "RMDlr Received")
	return
}

func pushToQueue(request ...DlrRequestInterface) {
	for _, req := range request {
		utils.RedisCon.Do("RPUSH", "dlrs", req.parseRequestString())
	}

	return
}

// ListenForDlrs on redis
func ListenForDlrs() {
	var dlrItem DlrRequest

	for {
		request, err := redis.Strings(utils.RedisCon.Do("BLPOP", "dlrs", 1))

		if err != nil && err == redis.ErrNil {
			time.Sleep(time.Second * 2)
		}

		for _, values := range request {
			if values != "dlrs" {
				err := json.Unmarshal([]byte(values), &dlrItem)
				if err != nil {
					log.Println("req Unmarshal", err)
				}
				saveDlr(&dlrItem)
			}
		}
	}
}

func saveDlr(req *DlrRequest) {

	recID, err := redis.String(utils.RedisCon.Do("GET", req.APIID))

	if err != nil && err == redis.ErrNil {
		log.Println("APIID Not Found:", req)
		if req.Retries >= 4 {
			saveHangingDlr(req)
		} else {
			req.Retries++
			utils.ScheduleTask("dlr_sched", req.parseRequestString(), req.Retries*5*60)
		}
		return
	}

	utils.RedisCon.Do("DEL", req.APIID)

	stmt, err := utils.DBCon.Prepare("insert into bsms_dlrstatus (status, reason, api_time, recipient_id) values (?, ?, ?, ?)")
	if err != nil {
		log.Println("Prepare Insert: ", err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(strings.ToUpper(req.Status), req.Reason, req.TimeReceived, recID)

	if err != nil {
		log.Println("Exec Insert: ", err)
		return
	}
	log.Println("Dlr saved: ", req.APIID)
	return
}

func saveHangingDlr(req *DlrRequest) {
	stmt, err := utils.DBCon.Prepare("insert into bsms_hangingdlrs (status, reason, api_time) values (?, ?, ?)")
	if err != nil {
		log.Println("Prepare hanging dlr Insert: ", err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(strings.ToUpper(req.Status), req.Reason, req.TimeReceived)

	if err != nil {
		log.Println("Exec hangingdlr Insert: ", err)
		return
	}
	log.Println("HangingDlr saved: ", req.APIID)
}

// PushToQueue requeue scheduled dlrs
func PushToQueue() {

	for {
		if _, err := utils.RedisCon.Do("WATCH", "dlr_sched"); err != nil {
			log.Println("WATCH error: ", err)
			return
		}

		// delta := time.Now().Add(time.Hour * 3).Unix()
		delta := time.Now().Unix()

		tasks, err := redis.Strings(
			utils.RedisCon.Do("ZRANGEBYSCORE", "dlr_sched", 0, delta))

		if err != nil {
			log.Println("sched Get error: ", err)
			return
		}

		if len(tasks) > 0 {
			for _, task := range tasks {
				utils.RedisCon.Send("MULTI")
				utils.RedisCon.Send("RPUSH", "dlrs", task)
				utils.RedisCon.Send("ZREM", "dlrs_sched", task)
				_, err := utils.RedisCon.Do("EXEC")
				if err != nil {
					log.Fatal("exec error: ", err)
				}
			}
		} else {
			if _, err := utils.RedisCon.Do("UNWATCH"); err != nil {
				log.Println("UNWATCH error: ", err)
			}
			time.Sleep(time.Second * 10)
		}
	}
}
