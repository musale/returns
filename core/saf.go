package pkg

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
var SMSNotifs = make(chan NotifySMSEnvelope, 100)

// NotifySMS handles sms notifications
func SafNotifyPage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ioutil readall: ", err)
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		w.Write([]byte("read all error"))
		return
	}

	var req NotifySMSEnvelope
	if err := xml.Unmarshal(body, &req); err != nil {
		log.Println("Xml unmarshal: ", err)
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		w.Write([]byte("xml unmarshal err"))
		return
	}

	log.Println("NotifySMS::req ", req)
	go func() {
		SMSNotifs <- req
	}()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
	w.Write([]byte(NotifySMSResp))
	return
}
