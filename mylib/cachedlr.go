package mylib

import "net/http"

func CacheDlrPage(w http.ResponseWriter, r *http.Request) {

	// get form parameters

	// params: api_id, message_id

	// save them as redis strings in async

	// SETEX <apiid> <2 wks> mid\

	// return json response
}
