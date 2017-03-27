package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/etowett/returns/utils"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func CacheDlrPage(w http.ResponseWriter, r *http.Request) {
	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	// Todo: Print all POST requests received

	APIID := r.FormValue("api_id")
	recID := r.FormValue("recipient_id")

	log.Println("Cache DLR request: api_id: ", APIID, " rec_id: ", recID)

	if _, err := redisCon.Do("SETEX", APIID, 1209600000000000, recID); err != nil {
		log.Fatal("cache error ", err)
	}

	json.NewEncoder(w).Encode(Response{
		Status: "success", Message: "dlr received",
	})

}
