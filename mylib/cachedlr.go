package mylib

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/etowett/returns/common"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func CacheDlrPage(w http.ResponseWriter, r *http.Request) {

	c := common.RedisPool().Get()
	defer c.Close()

	APIID := r.FormValue("api_id")
	recID := r.FormValue("recipient_id")

	ttl := int(time.Second * 60 * 60 * 24 * 14)

	log.Println("api_id: ", APIID, " rec_id: ", recID, " ttl: ", ttl)

	if _, err := c.Do("SETEX", APIID, ttl, recID); err != nil {
		log.Fatal("cache error ", err)
	}

	json.NewEncoder(w).Encode(Response{
		Status: "success", Message: "Normal received",
	})

}
