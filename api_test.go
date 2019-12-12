package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type Success struct {
	Key    string `json: "key"`
	Status string `json: "status"`
	Action string `json: "action"`
}

func TestKeyHandlerNewKey(t *testing.T) {
	uri := "/raspberry/keys/1234"
	method := "POST"
	sampleKey := createSampleSession()
	body, _ := json.Marshal(&sampleKey)

	recorder := httptest.NewRecorder()
	param := make(url.Values)
	req, err := http.NewRequest(method, uri+param.Encode(), strings.NewReader(string(body)))

	if err != nil {
		t.Fatal(err)
	}

	keyHandler(recorder, req)

	newSuccess := Success{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &newSuccess)

	if err != nil {
		t.Error("Could not unmarshal success message:\n", err)
	} else {
		if newSuccess.Status != "ok" {
			t.Error("Key not created, status error:\n", recorder.Body.String())
		}
		if newSuccess.Action != "added" {
			t.Error("Response is incorrect - action is not 'added :\n", recorder.Body.String())
		}
	}
}

func TestKeyHandlerUpdateKey(t *testing.T) {
	uri := "/raspberry/keys/1234"
	method := "PUT"
	sampleKey := createSampleSession()
	body, _ := json.Marshal(sampleKey)

	recorder := httptest.NewRecorder()
	param := make(url.Values)
	req, err := http.NewRequest(method, uri+param.Encode(), strings.NewReader(string(body)))

	if err != nil {
		t.Fatal(err)
	}

	keyHandler(recorder, req)

	newSuccess := Success{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &newSuccess)

	if err != nil {
		t.Error("Could not unmarshal success message:\n", err)
	} else {
		if newSuccess.Status != "ok" {
			t.Error("Key not created, status error:\n", recorder.Body.String())
		}
		if newSuccess.Action != "modified" {
			t.Error("Response is incorrect - action is not 'modified' :\n", recorder.Body.String())
		}
	}
}

func createKey() {
	uri := "/raspberry/keys/1234"
	method := "POST"
	sampleKey := createSampleSession()
	body, _ := json.Marshal(&sampleKey)

	recorder := httptest.NewRecorder()
	param := make(url.Values)
	req, _ := http.NewRequest(method, uri+param.Encode(), strings.NewReader(string(body)))

	keyHandler(recorder, req)
}

func TestKeyHandlerDeleteKey(t *testing.T) {
	createKey()

	uri := "/raspberry/keys/1234"
	method := "DELETE"

	recorder := httptest.NewRecorder()
	param := make(url.Values)
	req, err := http.NewRequest(method, uri+param.Encode(), nil)

	if err != nil {
		t.Fatal(err)
	}

	keyHandler(recorder, req)

	newSuccess := Success{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &newSuccess)

	if err != nil {
		t.Error("Could not unmarshal success message:\n", err)
	} else {
		if newSuccess.Status != "ok" {
			t.Error("Key not created, status error:\n", recorder.Body.String())
		}
		if newSuccess.Action != "deleted" {
			t.Error("Response is incorrect - action is not 'deleted' :\n", recorder.Body.String())
		}
	}
}

func TestCreatKeyHandlerCreateNewKey(t *testing.T) {
	createKey()

	uri := "/raspberry/keys/1234"
	method := "POST"
	sampleKey := createSampleSession()
	body, _ := json.Marshal(&sampleKey)

	recorder := httptest.NewRecorder()
	param := make(url.Values)
	req, err := http.NewRequest(method, uri+param.Encode(), strings.NewReader(string(body)))

	if err != nil {
		t.Fatal(err)
	}

	createKeyHandler(recorder, req)

	newSuccess := Success{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &newSuccess)

	if err != nil {
		t.Error("Could not unmarshal success message:\n", err)
	} else {
		if newSuccess.Status != "ok" {
			t.Error("Key not created, status error:\n", recorder.Body.String())
		}
		if newSuccess.Action != "create" {
			t.Error("Response is incorrect - action not 'create' :\n", recorder.Body.String())
		}
	}
}
