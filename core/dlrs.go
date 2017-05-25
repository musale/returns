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
type DLRRequest struct {
	APIID, Status, Reason string
	TimeReceived          time.Time
	Retries               int64
}

// DlrRequestInterface definition
type DlrRequestInterface interface {
	parseRequestString() string
	parseRequestMap() map[string]string
}

// Function to parse a DLRRequest item to string
func (request *DLRRequest) parseRequestString() string {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		log.Println("Dlr request to json string error", err)
	}
	return string(requestJSON)
}

// Function to parse a DLRRequest item to map[string]string
func (request *DLRRequest) parseRequestMap() map[string]string {
	return map[string]string{
		"api_id": request.APIID,
		"status": request.Status,
		"reason": request.Reason,
	}
}

var DLRReqChan = make(chan DLRRequest, 100)

// ATDlrPage rendering
func ATDlrPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("err: ATParseForm: ", err)
	}
	log.Println("ATDlrPage: ", r.Form)

	apiID := r.FormValue("id")
	apiStatus := r.FormValue("status")

	if strings.ToUpper(apiStatus) == "SUCCESS" {
		apiStatus = "DELIVRD"
	}

	request := DLRRequest{
		APIID: apiID, Status: strings.ToUpper(apiStatus),
		TimeReceived: time.Now(), Retries: 0,
	}

	if apiStatus == "Failed" || apiStatus == "Rejected" {
		request.Reason = r.FormValue("failureReason")
	}

	go func() {
		DLRReqChan <- request
	}()

	w.WriteHeader(200)
	_, err = fmt.Fprintf(w, "ATDlr Received")
	if err != nil {
		log.Println("err: ATWrite back resp: ", err)
	}
	return
}

// RMDlrPage rendering
func RMDlrPage(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Println("err: RMParseForm: ", err)
	}
	log.Println("RMDlrPage: ", r.Form)

	apiID := r.FormValue("sMessageId")
	apiStatus := r.FormValue("sStatus")

	request := DLRRequest{
		APIID: apiID, Status: strings.ToUpper(apiStatus),
		TimeReceived: time.Now(), Retries: 0,
	}

	if request.Status == "UNDELIV" {
		request.Status = "FAILED"
		request.Reason = "DeliveryFailure"
	}

	go func() {
		DLRReqChan <- request
	}()

	w.WriteHeader(200)
	_, err = fmt.Fprintf(w, "RMDlr Received")
	if err != nil {
		log.Println("err: RMWrite back resp: ", err)
	}
	return
}

// SafDlrPage rendering
func SafDlrPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("err: ATParseForm: ", err)
	}
	log.Println("ATDlrPage: ", r.Form)

	apiID := r.FormValue("message_id")
	phoneNumber := r.FormValue("number")
	apiStatus := r.FormValue("status")

	if strings.ToUpper(apiStatus) == "DeliveredToTerminal" {
		apiStatus = "DELIVRD"
	}

	request := DLRRequest{
		APIID: apiID, Status: phoneNumber + ":" + apiStatus,
		TimeReceived: time.Now(), Retries: 0,
	}

	go func() {
		DLRReqChan <- request
	}()

	w.WriteHeader(200)
	_, err = fmt.Fprintf(w, "ATDlr Received")
	if err != nil {
		log.Println("err: ATWrite back resp: ", err)
	}
	return
}

func QueueDlr(request ...DlrRequestInterface) error {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()
	for _, req := range request {
		if _, err := redisCon.Do("RPUSH", "dlrs", req.parseRequestString()); err != nil {
			log.Println("QueueDlr: ", err)
			return err
		}
	}
	return nil
}

// ListenForDlrs on redis
func ListenForDlrs() {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	var dlrItem DLRRequest
	for {
		dlrReq, err := redis.Strings(redisCon.Do("BLPOP", "dlrs", 1))

		if err != nil {
			if err == redis.ErrNil {
				time.Sleep(time.Second * 1)
			} else {
				log.Println("redisError: ", err)
			}
		}
		for _, dlrVal := range dlrReq {
			if dlrVal != "dlrs" {
				err := json.Unmarshal([]byte(dlrVal), &dlrItem)
				if err != nil {
					log.Println("req Unmarshal", err)
				}
				err = saveDlr(&dlrItem)
				if err != nil {
					log.Println("saveDlr error: ", err)
				}
			}
		}
	}
}

func saveDlr(req *DLRRequest) error {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	recID, err := redis.String(redisCon.Do("GET", req.APIID))

	if err != nil {
		if err == redis.ErrNil {
			if req.Retries > 8 {
				log.Println("Save Hanging DLR:", req)
				err = saveHangingDlr(req)
				if err != nil {
					return err
				}
			} else {
				log.Println("Sched DLR for retry:", req)
				req.Retries++
				err = utils.ScheduleTask(
					"dlr_sched", req.parseRequestString(), 5*60)
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
		return nil
	}

	if _, err = redisCon.Do("DEL", req.APIID); err != nil {
		return err
	}

	stmt, err := utils.DBCon.Prepare(
		"insert into bsms_dlrstatus (status, reason, api_time, " +
			"recipient_id) values (?, ?, ?, ?)",
	)
	if err != nil {
		log.Println("Prepare bsms_dlrstatus Insert")
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		strings.ToUpper(req.Status), req.Reason, req.TimeReceived, recID,
	)

	if err != nil {
		log.Println("Exec bsms_dlrstatus Insert")
		return err
	}
	log.Println("Dlr saved: ", req.APIID)
	return nil
}

func saveHangingDlr(req *DLRRequest) error {
	stmt, err := utils.DBCon.Prepare(
		"insert into bsms_hangingdlrs (api_id, status, reason, " +
			"api_time, insert_time) values (?, ?, ?, ?, ?)",
	)
	if err != nil {
		log.Println("Prepare hanging dlr Insert")
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		req.APIID, strings.ToUpper(req.Status), req.Reason, req.TimeReceived,
		time.Now(),
	)

	if err != nil {
		log.Println("Exec hangingdlr Insert")
		return err
	}
	log.Println("HangingDlr saved: ", req.APIID)
	return nil
}

// PushToQueue requeue scheduled dlrs
func PushToQueue() {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	for {
		if _, err := redisCon.Do("WATCH", "dlr_sched"); err != nil {
			log.Println("WATCH error: ", err)
			return
		}

		processTime := time.Now().Unix()

		tasks, err := redis.Strings(
			redisCon.Do("ZRANGEBYSCORE", "dlr_sched", 0, processTime))

		if err != nil {
			log.Println("sched Get error: ", err)
			return
		}

		if len(tasks) > 0 {
			for _, task := range tasks {
				redisCon.Send("MULTI")
				redisCon.Send("RPUSH", "dlrs", task)
				redisCon.Send("ZREM", "dlr_sched", task)
				_, err := redisCon.Do("EXEC")
				if err != nil {
					log.Fatal("exec error: ", err)
				}
			}
		} else {
			if _, err := redisCon.Do("UNWATCH"); err != nil {
				log.Println("UNWATCH error: ", err)
			}
			time.Sleep(time.Second * 10)
		}
	}
}
