package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/etowett/returns/common"
	"github.com/garyburd/redigo/redis"
)

// DlrRequest struct
type DlrRequest struct {
	APIID, Status, Reason string
	TimeReceived          time.Time
}

// DlrRequestInterface definition
type DlrRequestInterface interface {
	parseRequestString() string
	parseRequestMap() map[string]string
}

// Function to parse a DlrRequest item to string
func (request *DlrRequest) parseRequestString() string {
	requestJSON, _ := json.Marshal(request)
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
	if r.Method != common.POST {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

	aid := r.FormValue("id")
	status := r.FormValue("status")

	request := DlrRequest{
		APIID: aid, Status: status, TimeReceived: time.Now(),
	}

	if status == "Failed" || status == "Rejected" {
		request.Reason = r.FormValue("failureReason")
	}

	common.Logger.Println("ATDLR Request:", request)

	go pushToQueue(&request)

	fmt.Fprintf(w, "ATDlr Received")
	return
}

// DlrPage rendering
func RMDlrPage(w http.ResponseWriter, r *http.Request) {

	if r.Method != common.POST {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

	aid := r.FormValue("sMessageId")
	status := r.FormValue("sStatus")

	request := DlrRequest{
		APIID: aid, Status: status, TimeReceived: time.Now(),
	}

	common.Logger.Println("RMDLR Request:", request)

	go pushToQueue(&request)

	fmt.Fprintf(w, "RMDlr Received")
	return
}

func pushToQueue(request ...DlrRequestInterface) {
	pool := common.RedisPool().Get()
	defer pool.Close()

	for _, req := range request {
		pool.Do("RPUSH", "dlr_at", req.parseRequestString())
	}

	return
}

// ListenForDlrs on redis
func ListenForDlrs() {
	pool := common.RedisPool().Get()
	defer pool.Close()

	var dlrItem DlrRequest

	for {
		request, err := redis.Strings(pool.Do("BLPOP", "dlr_at", 1))

		if err != nil && err == redis.ErrNil {
			time.Sleep(time.Second * 2)
		}

		for _, values := range request {
			if values != "dlr_at" {
				err := json.Unmarshal([]byte(values), &dlrItem)
				if err != nil {
					common.Logger.Println("req Unmarshal", err)
				}
				saveDlr(dlrItem)
				// updateDlr(dlrItem)
			}
		}
	}
}

func saveDlr(req DlrRequest) {
	// get from redis string where api_id
	pool := common.RedisPool().Get()
	defer pool.Close()

	recID, err := redis.String(pool.Do("GET", req.APIID))

	if err != nil && err == redis.ErrNil {
		common.Logger.Println("APIID Not Found:", req)
		return
	}

	pool.Do("DEL", req.APIID)

	stmt, err := common.DbCon.Prepare("insert into bsms_dlrstatus (status, reason, api_time, recipient_id) values (?, ?, ?, ?)")
	if err != nil {
		common.Logger.Println("Prepare Insert: ", err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(strings.ToUpper(req.Status), req.Reason, req.TimeReceived, recID)

	if err != nil {

		common.Logger.Println("Exec Insert: ", err)
		return
	}
	common.Logger.Println("Dlr saved: ", req.APIID)
	return
}

func updateDlr(req DlrRequest) {
	stmt, err := common.DbCon.Prepare("update bsms_smsrecipient set status=?, reason=?, api_time=? where api_id=?")
	if err != nil {
		common.Logger.Fatal("Prepare Insert: ", err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(strings.ToUpper(req.Status), req.Reason, req.TimeReceived, req.APIID)

	if err != nil {
		common.Logger.Fatal("Exec Update: ", err)
		return
	}
	common.Logger.Println("Dlr saved: ", req.APIID)
	return
}
