package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

//APIRecipient struct
type APIRecipient struct {
	Number    string `json:"number"`
	Status    string `json:"status"`
	Cost      string `json:"cost"`
	MessageID string `json:"message_id"`
}

// MessageData struct
type MessageData struct {
	Message    string         `json:message`
	Recipients []APIRecipient `json:recipients`
}

// CostData struct
type CostData struct {
	Number string
	Status string
	Reason string
	APIID  string
	Cost   float64
}

// PushToAt pushes to the api
func PushToAt(to string, msg string, sid string) []APIRecipient {
	client := http.Client{}
	form := url.Values{}
	form.Add("username", "etowett")
	form.Add("message", msg)
	form.Add("to", to)
	form.Add("from", sid)
	req, err := http.NewRequest(
		"POST", os.Getenv("AFT_URL"), strings.NewReader(form.Encode()))

	if err != nil {
		fmt.Println("Request Error: ", err)
		return []APIRecipient{}
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	req.Header.Add("Accept", "Application/json")
	req.Header.Add("apikey", "abcd1234")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Do Error: ", err)
		return []APIRecipient{}
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("AFT: ", string(body))

	retData := make(map[string]MessageData)

	err = json.Unmarshal(body, &retData)

	if err != nil {
		fmt.Println("Marshal Error: ", err)
		return []APIRecipient{}
	}

	return retData["SMSMessageData"].Recipients
}

// GetCosts gets the cost
func GetCosts(recs []APIRecipient, costs map[string]float64) ([]CostData, string) {
	var rData []CostData
	mcost := 0.00
	for _, rec := range recs {
		nrec := CostData{
			Number: rec.Number, Reason: "", Cost: 0.00, APIID: "",
		}
		if rec.Status == "Success" {
			nrec.Status = "sent"
			nrec.APIID = rec.MessageID
			nrec.Cost = costs[rec.Number]
		} else if rec.Status == "User In BlackList" {
			nrec.Status = "opted_out"
			nrec.Reason = "User Opted Out"
		} else if rec.Status == "Invalid Phone Number" {
			nrec.Status = "invalid_num"
			nrec.Reason = "Number Invalid"
		} else if rec.Status == "Could Not Send" {
			nrec.Status = FAILED
			nrec.Reason = "Rejected"
		} else if rec.Status == "Insufficient Balance" {
			nrec.Status = FAILED
			nrec.Reason = "Insufficient Balance"
		} else {
			nrec.Status = FAILED
			nrec.Reason = rec.Status
		}
		// cost, _ := strconv.ParseFloat(rec["cost"], 64)
		rData = append(rData, nrec)
		mcost += nrec.Cost
	}
	return rData, fmt.Sprintf("%.2f", mcost)
}

// RedisPool returns a redis pool
func RedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// ScheduleTask creates a schedule for future
func ScheduleTask(queue string, data string, delay int64) {
	redisCon := RedisPool().Get()
	defer redisCon.Close()

	runAt := time.Now().Unix() + delay

	redisCon.Do("ZADD", queue, runAt, data)

	return
}
