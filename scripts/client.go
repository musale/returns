package main

import (
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
	log.Println("Inbox: ", resps)
	return
}

func sendInbox() (string, error) {
	url := "http://127.0.0.1/inbox"
	form := url.Values{}
	form.Add("from", getPhone())
	form.Add("to", getShortCode())
	form.Add("text", getMessage())
	form.Add("date", time.Now())
	form.Add("id", time.Now().Unix())

	return makeRequest(url, form), nil
}

func getPhone() string {
	return "254727372285"
}

func getShortCode() string {
	return "31390"
}

func getMessage() string {
	return "Hello world!"
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
