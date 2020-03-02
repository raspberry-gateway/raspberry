package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Test struct {
	Functions []string `json:"functions"`
	It        string   `json:"it"`
}

func TestFlattenBaseCase(t *testing.T) {
	data := make(map[string]interface{})

	// puts fields
	data["app"] = "raspberry"
	data["version"] = "0.0.1"

	// puts map
	group := make(map[string]interface{})
	group["ID"] = 100
	group["desc"] = "parent group"

	subGroup := make([]string, 3, 3)
	subGroup[0] = "a"
	subGroup[1] = "b"
	subGroup[2] = "c"
	group["subGroup"] = subGroup

	data["group"] = group

	// put struct
	test := Test{It: "test reflect struct", Functions: []string{"f1", "f2"}}
	group["test"] = test

	flatMap, err := Flatten(data)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "raspberry", flatMap["app"], "flatten function failure")
	assert.Equal(t, "0.0.1", flatMap["version"], "flatten function failure")
	assert.Equal(t, "a", flatMap["group.subGroup.0"], "flatten array/slice failure")
	assert.Equal(t, "b", flatMap["group.subGroup.1"], "flatten array/slice failure")
	assert.Equal(t, "100", flatMap["group.ID"], "flatten map failure")
	assert.Equal(t, "f2", flatMap["group.test.Functions.1"], "flatten struct failure")
}
