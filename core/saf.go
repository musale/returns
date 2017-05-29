package core

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/etowett/returns/utils"
)

// NotifyEnvelope is the struct that creates the xml payload
type NotifyEnvelope struct {
	NotifyHeader struct {
		SOAPHeader struct {
			TimeStamp  string `xml:"timeStamp"`
			SubReqID   string `xml:"subReqID"`
			UniqueueID string `xml:"traceUniqueID"`
		} `xml:"NotifySOAPHeader"`
	} `xml:"Header"`
	NotifyBody struct {
		DLRReceipt struct {
			Correlator string `xml:"correlator"`
			DLRStatus  struct {
				Number string `xml:"address"`
				Status string `xml:"deliveryStatus"`
			} `xml:"deliveryStatus"`
		} `xml:"notifySmsDeliveryReceipt"`
	} `xml:"Body"`
}

// Tel is the first 4 characters before a telephone number
var Tel = "tel:"

// NotifyResp is the xml payload for notifySmsDeliveryReceiptResponse
var NotifyResp = `<?xml version="1.0" encoding="UTF-8"?><soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loc="http://www.csapi.org/schema/parlayx/sms/notification/v2_2/local"><soapenv:Header /><soapenv:Body><loc:notifySmsDeliveryReceiptResponse /></soapenv:Body></soapenv:Envelope>`

// NotifySMS handles sms notifications
func SafNotifyPage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ioutil readall: ", err)
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		w.Write([]byte("read all error"))
		return
	}

	log.Println("SafNotifyPage: ", string(body))

	var req NotifyEnvelope
	if err := xml.Unmarshal(body, &req); err != nil {
		log.Println("Xml unmarshal: ", err)
		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		w.Write([]byte("xml unmarshal err"))
		return
	}

	apiID := req.NotifyBody.DLRReceipt.Correlator
	phoneNumber := req.NotifyBody.DLRReceipt.DLRStatus.Number
	apiStatus := req.NotifyBody.DLRReceipt.DLRStatus.Status
	if phoneNumber[0:4] == Tel {
		phoneNumber = phoneNumber[4:]
	}

	apiReason := ""
	xStatus := []string{"deliveredtoterminal", "deliveredtonetwork"}

	if utils.InArray(strings.ToLower(apiStatus), xStatus) {
		apiStatus = "DELIVRD"
	} else {
		apiReason = apiStatus
		apiStatus = "FAILED"
	}

	request := DLRRequest{
		APIID: phoneNumber + ":" + apiID, Status: apiStatus,
		TimeReceived: time.Now(), Retries: 0,
	}

	if len(apiReason) > 1 {
		request.Reason = apiReason
	}

	go func(request *DLRRequest) {
		DLRReqChan <- request
	}(&request)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/xml; charset=UTF-8")
	w.Write([]byte(NotifyResp))
	return
}
