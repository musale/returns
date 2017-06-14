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
	// Todo: Print all POST requests received

	APIID := r.FormValue("api_id")
	recID := r.FormValue("recipient_id")

	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	log.Println("Cache DLR request: api_id: ", APIID, " rec_id: ", recID)

	if _, err := redisCon.Do("SETEX", APIID, 1209600000000000, recID); err != nil {
		log.Fatal("cache error ", err)
	}

	json.NewEncoder(w).Encode(Response{
		Status: "success", Message: "dlr received",
	})
	return
}

type CacheReq struct {
	RecID string `json:"rid"`
	APIID string `json:"api_id"`
}

func CacheBulkDlrPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("CacheBulk ParseFormError: ", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			Status: "not_ok", Message: "Problem with your request",
		})
		return
	}
	log.Println("CacheBulkDlrPage: ", r.Form)

	apiIDs := r.FormValue("ids")
	var allDlrs []CacheReq
	err = json.Unmarshal([]byte(apiIDs), &allDlrs)
	if err != nil {
		log.Println("Error Unmarshal:", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			Status: "not_ok", Message: "Problem with your request",
		})
		return
	}

	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	for _, rec := range allDlrs {
		if _, err := redisCon.Do(
			"SETEX", rec.APIID, 1209600000000000, rec.RecID); err != nil {
			log.Println("cache error ", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Response{
				Status: "not_ok", Message: "Problem with your request",
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Status: "success", Message: "dlr received",
	})
	return
}
