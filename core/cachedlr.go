package core

import (
	"encoding/json"
	"log"
	"net/http"

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

	// ttl := int(time.Second * 60 * 60 * 24 * 14)

	common.Logger.Println("Cache DLR request: api_id: ", APIID, " rec_id: ", recID)

	if _, err := c.Do("SETEX", APIID, 1209600000000000, recID); err != nil {
		log.Fatal("cache error ", err)
	}

	json.NewEncoder(w).Encode(Response{
		Status: "success", Message: "dlr received",
	})

}
