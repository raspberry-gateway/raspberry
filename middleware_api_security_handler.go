package main

import (
	"fmt"
	"net/http"
)

// CheckIsAPIOwner will ensure that the accessor of the Raspberry API has the correct security credentials - this is a
// shared secret between the client and the owner and is set in the raspberry.conf file. This should never be made public!
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
