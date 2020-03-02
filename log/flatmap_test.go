package log

import (
	"testing"
)

type Group struct {
	ID int64 `json: "ID"`
}

func TestFlattenBaseCase(t *testing.T) {
	data := make(map[string]interface{})

	// puts fields
	data["app"] = "raspberry"
	data["version"] = "0.0.1"

}
