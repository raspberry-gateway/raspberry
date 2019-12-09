package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ApiModifyKeySuccess struct {
	Key    string `json: "key"`
	Status string `json: "status"`
	Action string `json: "action"`
}

func handleAddOrUpdate(keyName string, r *http.Request) ([]byte, int) {
	success := true
	decoder := json.NewDecoder(r.Body)
	var newSession SessionState
	err := decoder.Decode(&newSession)
	code := 200

	if err != nil {
		log.Error("Couldn't decode new session object")
		log.Error(err)
		code = 403
		success = false
	} else {
		// Update our session object (create it)
		authManager.UpdateSession(keyName, newSession)
	}

	var responseMessage []byte
	var action string
	if r.Method == "POST" {
		action = "added"
	} else {
		action = "modified"
	}

	if success {
		response := ApiModifyKeySuccess{
			keyName,
			"ok",
			action}

		responseMessage, err = json.Marshal(&response)

		if err != nil {
			log.Error("Could not create response message")
			log.Error(err)
			code = 500
			responseMessage = []byte(systemError)
		}
	}

	return responseMessage, code
}

func keyHandler(w http.ResponseWriter, r *http.Request) {
	keyName := r.URL.Path[len("/raspberry/keys/"):]
	var responseMessage []byte
	var code int

	if r.Method == "POST" || r.Method == "PUT" {
		responseMessage, code = handleAddOrUpdate(keyName, r)
		w.WriteHeader(code)
		fmt.Fprintf(w, string(responseMessage))
	} else if r.Method == "GET" {
		if keyName != "" {
			// Return single key detail
			responseMessage, code = handleGetDetail(keyName)
		} else {
			// Return list of keys
			responseMessage, code = handleGetAllKeys()
		}
	} else if r.Method == "DELETE" {
		// Remove a key
		responseMessage, code = handleDeleteKey(keyName)
	} else {
		// Return Not suppored message (and code)
		code = 400
		responseMessage = []byte(systemError)
	}

	w.WriteHeader(code)
	fmt.Fprintf(w, string(responseMessage))
}

type APIStatusMessage struct {
	Status  bool   `json: "status"`
	Message string `json: "message"`
}

func handleGetDetail(sessionKey string) ([]byte, int) {
	success := true
	var responseMessage []byte
	var err error
	code := 200

	thisSession, ok := authManager.GetSessionDetail(sessionKey)
	if !ok {
		success = false
	} else {
		responseMessage, err = json.Marshal(&thisSession)
		if err != nil {
			log.Error("Marshaling failed")
			log.Error(err)
			success = false
		}
	}

	if !success {
		notFound := APIStatusMessage{false, "Key not found"}
		responseMessage, _ = json.Marshal(&notFound)
		code = 404
	}
	return responseMessage, code
}

type APIAllKeys struct {
	ApiKeys []string `json: "api_keys"`
}

func handleGetAllKeys() ([]byte, int) {
	success := true
	var responseMessage []byte
	code := 200

	var err error

	sessions := authManager.GetSessions()
	sessionsObj := APIAllKeys{sessions}

	responseMessage, err = json.Marshal(&sessionsObj)
	if err != nil {
		log.Error("Marshaling failed")
		log.Error(err)
		success = false
		code = 500
	}

	if !success {
		return []byte(systemError), code
	}
	return responseMessage, code
}

func handleDeleteKey(keyName string) ([]byte, int) {
	var responseMessage []byte
	var err error
	authManager.Store.DeleteKey(keyName)
	code := 200

	statusObj := APIStatusMessage{true, ""}
	responseMessage, err = json.Marshal(&statusObj)

	if err != nil {
		log.Error("Marshaling falied")
		log.Error(err)
		code = 500
		return []byte(systemError), code
	}
	return responseMessage, code
}
