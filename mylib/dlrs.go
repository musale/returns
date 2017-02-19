package mylib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/etowett/returns/common"
	"github.com/garyburd/redigo/redis"
)

// DlrRequest struct
type DlrRequest struct {
	APIID, Status, Reason string
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

// DlrPage rendering
func DlrPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != common.POST {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

	aid := r.FormValue("id")
	status := r.FormValue("status")
	reason := r.FormValue("reason")

	request := DlrRequest{APIID: aid, Status: status}

	if status == "Failed" {
		request.Reason = reason
	}

	go pushToQueue(&request)

	fmt.Fprintf(w, "Dlr Received")
	return
}

func pushToQueue(requests ...DlrRequestInterface) {
	pool := common.RedisPool().Get()
	defer pool.Close()

	for _, request := range requests {
		pool.Do("RPUSH", "dlr_at", request.parseRequestString())
	}
}

// ListenForDlrs on redis
func ListenForDlrs() {
	logger := common.Logger
	pool := common.RedisPool().Get()
	defer pool.Close()
	var dlrItem DlrRequest
	for {
		request, err := redis.Strings(pool.Do("BLPOP", "dlr_at", 1))
		if err != nil {
			logger.Println("ERROR POPPING DLR:: ", err)
			time.Sleep(time.Second * 2)
		}

		for _, values := range request {
			byteValue := []byte(values)
			err := json.Unmarshal(byteValue, &dlrItem)
			if err != nil {
				logger.Println(err)
			}

			saveDlr(dlrItem.parseRequestMap())
		}
	}
}

func saveDlr(req map[string]string) {
	logger := common.Logger
	db := common.DbCon
	stmt, err1 := db.Prepare("update bsms_smsrecipient set status=?, reason=? where api_id=?")
	if err1 != nil {
		logger.Fatal("Couldn't prepare for dlr update", err1)
		return
	}

	defer stmt.Close()

	_, err := stmt.Exec(req["status"], req["reason"], req["api_id"])

	if err != nil {
		logger.Println("Cannot run update dlr", err)
		return
	}
	logger.Println("Dlr saved: ", req["api_id"])
	return
}
