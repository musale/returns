package core

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/etowett/returns/utils"
)

const (
	notifySMSResp = `<?xml version="1.0" encoding="UTF-8"?><soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loc="http://www.csapi.org/schema/parlayx/sms/notification/v2_2/local"><soapenv:Header /><soapenv:Body><loc:notifySmsReceptionResponse /></soapenv:Body></soapenv:Envelope>`
)

// NotifySMSEnvelope is the structure of an SMS notification response
type NotifySMSEnvelope struct {
	NotifyHeader struct {
		SOAPHeader struct {
			RevID       string `xml:"spRevId"`
			RevPassword string `xml:"spRevpassword"`
			SpID        string `xml:"spId"`
			ServiceID   string `xml:"serviceId"`
			LinkID      string `xml:"linkid"`
			TransID     string `xml:"traceUniqueID"`
		} `xml:"NotifySOAPHeader"`
	} `xml:"Header"`
	NotifyBody struct {
		SMSReception struct {
			Correlator string `xml:"correlator"`
			Message    struct {
				Message          string `xml:"message"`
				Number           string `xml:"senderAddress"`
				ActivationNumber string `xml:"smsServiceActivationNumber"`
				Date             string `xml:"dateTime"`
			} `xml:"message"`
		} `xml:"notifySmsReception"`
	} `xml:"Body"`
}

// SMSNotifs is the channel handling at most 100 sms notifications
var SMSNotifs = make(chan *NotifySMSEnvelope, 100)

// SafInboxPage rendering
func SafInboxPage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ioutil readall: ", err)
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		w.Write([]byte("read all error"))
		return
	}

	log.Println("SafInboxPage:", string(body))
	var req NotifySMSEnvelope
	if err := xml.Unmarshal(body, &req); err != nil {
		log.Println("Xml unmarshal: ", err)
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		w.Write([]byte("xml unmarshal err"))
		return
	}

	go func(notReq *NotifySMSEnvelope) {
		SMSNotifs <- notReq
	}(&req)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
	w.Write([]byte(notifySMSResp))
	return
}

// ProcessSMSNofit processes the SMS notification
func (req *NotifySMSEnvelope) ProcessSMSNofit() error {
	// revID := req.NotifyHeader.SOAPHeader.RevID
	// revPass := req.NotifyHeader.SOAPHeader.RevPassword
	// linkID := req.NotifyHeader.SOAPHeader.LinkID
	transID := req.NotifyHeader.SOAPHeader.TransID
	// correlator := req.NotifyBody.SMSReception.Correlator
	message := req.NotifyBody.SMSReception.Message.Message
	telNumber := req.NotifyBody.SMSReception.Message.Number
	actNum := req.NotifyBody.SMSReception.Message.ActivationNumber
	dateTime := req.NotifyBody.SMSReception.Message.Date

	if telNumber[0:4] == "tel:" {
		telNumber = telNumber[4:]
	}

	stmt, err := utils.DBCon.Prepare("insert into bsms_smsinbox(is_read, sender, short_code, api_id, message, user_id, deleted, api_date, insert_date) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		log.Println("Couldn't prepare for inbox insert", err)
		return err
	}

	res, err := stmt.Exec(
		0, telNumber, actNum, transID, message, 1, 0,
		dateTime, time.Now(),
	)
	if err != nil {
		log.Println("Cannot run insert Inbox", err)
		return err
	}

	oid, _ := res.LastInsertId()
	log.Println("Saved Inbox, id:", oid)
	return nil
}
