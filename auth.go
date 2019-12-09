package main

import (
	"encoding/json"
)

// AuthorisationHandler is used to validate a session key,
// implementing IsKeyAuthorised() to validate if a key exists or
// is valid in any way (e.g. cryptographic signing etc.). Returns
// a SessionState object (deserialised JSON)
type AuthorisationHandler interface {
	IsKeyAutorised(string) (bool, SessionState)
}

// AuthorisationManager implements AuthorisationHandler,
// requires a StorageHandler to interact with key store
type AuthorisationManager struct {
	Store StorageHandler
}

// IsKeyAuthorised checks if key exists and can be read into a SessionState object
func (b AuthorisationManager) IsKeyAuthorised(keyName string) (bool, SessionState) {
	jsonKeyVal, err := b.Store.GetKey(keyName)
	var newSession SessionState
	if err != nil {
		log.Warning("Invalid key detected, not found in storage engine")
		return false, newSession
	}
	err = json.Unmarshal([]byte(jsonKeyVal), &newSession)
	if err != nil {
		log.Error("Couldn't unmarshal session object")
		log.Error(err)
		return false, newSession
	}
	return true, newSession
}

// UpdateSession updates the session in the storage engine
func (b AuthorisationManager) UpdateSession(keyName string, session SessionState) {
	v, _ := json.Marshal(session)
	b.Store.SetKey(keyName, string(v))
}

func (b AuthorisationManager) GetSessionDetail(keyName string) (SessionState, bool) {
	jsonKeyVal, err := b.Store.GetKey(keyName)
	var newSession SessionState
	if err != nil {
		log.Warning("key does not exist")
		return newSession, false
	}
	err = json.Unmarshal([]byte(jsonKeyVal), &newSession)
	if err != nil {
		log.Error("Couldn't unmarshal session object")
		log.Error(err)
		return newSession, false
	}
	return newSession, true
}

func (b AuthorisationManager) GetSessions() []string {
	return b.Store.Getkeys()
}
