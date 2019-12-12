package main

// KeyError is a standard error for when a key is not found in the storage engine
type KeyError struct{}

func (e KeyError) Error() string {
	return "Key not found"
}

// StorageHandler is a standard interface to a storage backend,
// used by AuthorisationManager to read and write key values to the backend
type StorageHandler interface {
	GetKey(string) (string, error) // Returned string is expected to be a JSON object (SessionState)
	SetKey(string, string)         // Second input string is expected to be a JSON object (SessionState)
	GetKeys() []string
	DeleteKey(string) bool
}

// InMemoryStorageManager implements the StorageHandler interface,
// it uses an in-momery map to store sessions, should only be used
// for testing purposes
type InMemoryStorageManager struct {
	Sessions map[string]string
}

// GetKey retrives the key from the in-memory map
func (s InMemoryStorageManager) GetKey(keyName string) (string, error) {
	value, ok := s.Sessions[keyName]
	if !ok {
		return "", KeyError{}
	}
	return value, nil
}

// SetKey updates the in-memory key
func (s InMemoryStorageManager) SetKey(keyName string, sessionState string) {
	s.Sessions[keyName] = sessionState
}

func (s InMemoryStorageManager) GetKeys() []string {
	sessions := make([]string, 0, len(s.Sessions))
	for key, _ := range s.Sessions {
		sessions = append(sessions, key)
	}

	return sessions
}

func (s InMemoryStorageManager) DeleteKey(keyName string) bool {
	delete(s.Sessions, keyName)
	return true
}
