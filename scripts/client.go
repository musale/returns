package main

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func main() {
	resp, err := sendInbox()
	if err != nil {
		log.Println("inbox error: ", err)
	}
	log.Println("Inbox: ", resp)
	return
}

func sendInbox() (string, error) {
	inboxURL := "http://callbacks.local/inbox"
	inID := getMD5Hash(time.Now().String())
	form := url.Values{}
	form.Add("from", getPhone())
	form.Add("to", getShortCode())
	form.Add("text", getMessage())
	form.Add("date", time.Now().String())
	form.Add("id", inID)

	response, err := makeRequest(inboxURL, form)
	if err != nil {
		return "", err
	}
	return response, nil
}

func getPhone() string {
	return "254727372285"
}

func getShortCode() string {
	return "31391"
}

func getMessage() string {
	return "This should go express!"
}

func makeRequest(
	destUrl string, form url.Values,
) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest(
		"POST", destUrl, strings.NewReader(form.Encode()))

	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
