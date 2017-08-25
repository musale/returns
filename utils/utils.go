package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
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

// Costs for used amount
type Costs struct {
	TotalCost float64    `json:"total_cost"`
	CostData  []CostData `json:"cost_data"`
}

// InArray checks if an item is in the array
func InArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// PushToAt common function to interface with API
func PushToAt(to string, msg string, sid string) MessageData {
	client := http.Client{}
	form := url.Values{}
	form.Add("username", os.Getenv("AFT_USER"))
	form.Add("message", msg)
	form.Add("to", to)
	form.Add("from", sid)
	req, err := http.NewRequest(
		"POST", os.Getenv("AFT_URL"), strings.NewReader(form.Encode()))

	if err != nil {
		log.Println("AFTError: ", err)
		return MessageData{"request error", []APIRecipient{}}
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	req.Header.Add("Accept", "Application/json")
	req.Header.Add("apikey", os.Getenv("AFT_KEY"))

	resp, err := client.Do(req)

	if err != nil {
		log.Println("AFTError: ", err)
		return MessageData{"Do error", []APIRecipient{}}
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("AFTError: ", err)
		return MessageData{"readall error", []APIRecipient{}}
	}

	log.Println("AFT: ", string(body))

	retData := make(map[string]MessageData)

	err = json.Unmarshal(body, &retData)

	if err != nil {
		log.Println("api resp marshall: ", err)
		return MessageData{"unmarshall error", []APIRecipient{}}
	}

	return retData["SMSMessageData"]
}

// GetCosts format API result to something I can use
func GetCosts(recs []APIRecipient, costs map[string]float64) Costs {
	var costData []CostData
	messageCost := 0.00
	for _, rec := range recs {
		recData := CostData{Number: rec.Number, Cost: 0.00, Status: "FAILED"}
		if rec.Status == "Success" {
			recData.Status = "SENT"
			recData.APIID = rec.MessageID
			recData.Cost = costs[rec.Number]
		} else if rec.Status == "User In BlackList" {
			recData.Status = "OPTED_OUT"
			recData.Reason = "User Opted Out"
		} else if rec.Status == "Invalid Phone Number" {
			recData.Status = "INVALID_NUM"
			recData.Reason = "Number Invalid"
		} else if rec.Status == "Could Not Send" {
			recData.Reason = "Failed to Send"
		} else if rec.Status == "Insufficient Balance" {
			recData.Reason = "Insufficient Balance"
		} else {
			recData.Reason = rec.Status
		}
		costData = append(costData, recData)
		messageCost += recData.Cost
	}
	return Costs{TotalCost: messageCost, CostData: costData}
}

// DummyCosts data when api fail to return
func DummyCosts(to string) Costs {
	var costData []CostData
	for _, number := range strings.Split(to, ",") {
		costData = append(costData, CostData{
			Number: number, Reason: "Could Not Send", Cost: 0.00, APIID: "",
			Status: "Failed",
		})
	}
	return Costs{TotalCost: 0.00, CostData: costData}
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
			_, err = c.Do("AUTH", os.Getenv("RED_PASS"))
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// ScheduleTask creates a schedule for future
func ScheduleTask(queue string, data string, delay int64) error {
	redisCon := RedisPool().Get()
	defer redisCon.Close()

	runAt := time.Now().Unix() + delay

	if _, err := redisCon.Do("ZADD", queue, runAt, data); err != nil {
		return err
	}

	return nil
}
