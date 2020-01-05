package main

import (
	"encoding/json"
	"time"
)

// AuthorisationHandler is used to validate a session key,
// implementing IsKeyAuthorised() to validate if a key exists or
// is valid in any way (e.g. cryptographic signing etc.). Returns
// a SessionState object (deserialised JSON)
type AuthorisationHandler interface {
	IsKeyAutorised(string) (bool, SessionState)
	IsKeyExpired(SessionState) bool
}

// AuthorisationManager implements AuthorisationHandler,
// requires a StorageHandler to interact with key store
type AuthorisationManager struct {
	Store StorageHandler
}

// IsKeyAuthorised checks if key exists and can be read into a SessionState object
func (b AuthorisationManager) IsKeyAuthorised(keyName string) (bool, SessionState) {
	jsonKeyVal, marshalErr := b.Store.GetKey(keyName)
	var newSession SessionState
	if marshalErr != nil {
		log.Warning("Invalid key detected, not found in storage engine")
		return false, newSession
	}

	if marshalErr = json.Unmarshal([]byte(jsonKeyVal), &newSession); marshalErr != nil {
		log.Error("Couldn't unmarshal session object")
		log.Error(marshalErr)
		return false, newSession
	}

	return true, newSession
}

// IsKeyExpired checks if a key has expired, if the value of SessionState.Expires is 0, it will be ignored
func (b AuthorisationManager) IsKeyExpired(newSession *SessionState) bool {
	if newSession.Expires >= 1 {
		diff := newSession.Expires - time.Now().Unix()
		if diff > 0 {
			return false
		}
		return true
	}
	return false
}

// UpdateSession updates the session in the storage engine
func (b AuthorisationManager) UpdateSession(keyName string, session SessionState) {
	v, _ := json.Marshal(session)
	keyExp := (session.Expires - time.Now().Unix()) + 300 // Add 5 minutes to key expiry, just in case
	b.Store.SetKey(keyName, string(v), keyExp)
}

// GetSessionDetail returns the session detail using the storage engine (either in memory or Redis)
func (b AuthorisationManager) GetSessionDetail(keyName string) (SessionState, bool) {
	jsonKeyVal, marshalErr := b.Store.GetKey(keyName)
	var newSession SessionState
	if marshalErr != nil {
		log.Warning("key does not exist")
		return newSession, false
	}

	if marshalErr = json.Unmarshal([]byte(jsonKeyVal), &newSession); marshalErr != nil {
		log.Error("Couldn't unmarshal session object")
		log.Error(marshalErr)
		return newSession, false
	}
	return newSession, true
}

// GetSessions returns all sessions in the key store that match a filter key (a prefix)
func (b AuthorisationManager) GetSessions(filter string) []string {
	return b.Store.GetKeys(filter)
}
