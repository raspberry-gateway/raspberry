package main

import (
	"fmt"
	"net/http"
)

func CheckIsAPIOwner(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		raspberryAuthKey := r.Header.Get("X-Raspberry-Authorisation")
		if raspberryAuthKey != config.Secret {
			// Error
			log.Warning("Attempted administractive access with invalid or missing key!")

			responseMessage := createError("Method not supported")
			w.WriteHeader(403)
			fmt.Fprintf(w, string(responseMessage))

			return
		}

		handler(w, r)
	}
}
