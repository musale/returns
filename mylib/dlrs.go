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
	parseRequest() string
}

// Function to parse a DlrRequest
func (request *DlrRequest) parseRequest() string {
	requestJSON, _ := json.Marshal(request)
	return string(requestJSON)
}

// DlrPage rendering
func DlrPage(w http.ResponseWriter, r *http.Request) {
	logger := common.Logger
	if r.Method != common.POST {
		fmt.Fprintf(w, "Method Not Allowed")
		return
	}

	aid := r.FormValue("id")
	status := r.FormValue("status")
	reason := r.FormValue("reason")

	// request := map[string]string{
	// 	"aid": aid, "status": status, "reason": reason,
	// }
	request := DlrRequest{APIID: aid, Status: status}

	if status == "Failed" {
		request.Reason = reason
	}

	go pushToQueue(&request)

	logger.Println("Dlr Request: ", request)

	fmt.Fprintf(w, "Dlr Received")
	return
}

func pushToQueue(requests ...DlrRequestInterface) {
	pool := common.RedisPool().Get()
	defer pool.Close()

	for _, request := range requests {
		pool.Do("RPUSH", "dlr_at", request.parseRequest())
	}
}

// ListenForDlrs on redis
func ListenForDlrs() {
	logger := common.Logger
	logger.Println("HEY! ", "hello")
	pool := common.RedisPool().Get()
	defer pool.Close()
	for {
		// dlrItem := make(map[string]DlrRequest)

		request, err := redis.Strings(pool.Do("BLPOP", "dlr_at", 1))
		logger.Println("REZ", request)
		if err != nil {
			logger.Println("ERROR POPPING DLR:: ", err)
			time.Sleep(time.Second * 2)
		}

		// TODO: JSON DECODE
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

	_, err := stmt.Exec(req["status"], req["reason"], req["aid"])

	if err != nil {
		logger.Println("Cannot run update dlr", err)
		return
	}
	logger.Println("Dlr saved: ", req["aid"])
	return
}
