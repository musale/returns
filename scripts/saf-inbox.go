package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	// NOTIFY_SMS_URL           = "http://192.168.122.95:8017/saf-inbox"
	NOTIFY_SMS_URL           = "http://callbacks.smsleopard.com/saf-inbox"
	NOTIFY_SMS_URL_REG_QUERY = `
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:loc="http://www.csapi.org/schema/parlayx/sms/notification/v2_2/local" xmlns:v2="http://www.huawei.com.cn/schema/common/v2_1">
    <soapenv:Header>
        <ns1:NotifySOAPHeader xmlns:ns1="http://www.huawei.com.cn/schema/common/v2_1">
            <ns1:spRevId>ekubai</ns1:spRevId>
            <ns1:spRevpassword>Abc123</ns1:spRevpassword>
            <ns1:spId>601399</ns1:spId>
            <ns1:serviceId>6013992000001491</ns1:serviceId>
            <ns1:linkid>07161722430758000009</ns1:linkid>
            <ns1:traceUniqueID>404090102591207161422430892004</ns1:traceUniqueID>
        </ns1:NotifySOAPHeader>
    </soapenv:Header>
    <soapenv:Body>
        <loc:notifySmsReception>
            <loc:correlator>123</loc:correlator>
            <loc:message>
                <message>This is a test message</message>
                <senderAddress>tel:722123456</senderAddress>
                <smsServiceActivationNumber>1234</smsServiceActivationNumber>
                <dateTime>2012-07-03T00:00:00.000+08:00</dateTime>
            </loc:message>
        </loc:notifySmsReception>
    </soapenv:Body>
</soapenv:Envelope>
`
)

func main() {

	httpClient := new(http.Client)
	resp, err := httpClient.Post(NOTIFY_SMS_URL, "text/xml; charset=utf-8", bytes.NewBufferString(NOTIFY_SMS_URL_REG_QUERY))
	if err != nil {
		log.Fatal("Post error: ", err)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Post error: ", err)
	}
	in := string(b)
	log.Println("Response: ", in)
	return
}
