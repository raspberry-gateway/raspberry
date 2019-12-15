package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/vmihailenco/msgpack"
)

// AnalyticsRecord encodes the details of a request
type AnalyticsRecord struct {
	Method        string
	Path          string
	ContentLength int64
	UserAgent     string
	Day           int
	Month         time.Month
	Year          int
	Hour          int
	ResponseCode  int
}

// AnalyticsError is an error for when writing to the storage engine fails
type AnalyticsError struct{}

func (e AnalyticsError) Error() string {
	return "Recording request failed!"
}

type AnalyticsHandler interface {
	RecordHit(AnalyticsRecord) error
	PurgeCache()
}

type RedisAnalyticsHandler struct {
	Store RedisStorageManager
}

func (r RedisAnalyticsHandler) RecordHit(thisRecord AnalyticsRecord) error {
	encoded, err := msgpack.Marshal(thisRecord)
	u5, _ := uuid.NewV4()

	keyName := fmt.Sprintf("%d%d%d%d-%s", thisRecord.Year, thisRecord.Month, thisRecord.Day, thisRecord.Hour, u5.String())

	if err != nil {
		log.Error("Error encoding analytics data:")
		log.Error(err)
		return AnalyticsError{}
	}
	r.Store.SetKey(keyName, string(encoded))
	return nil
}

func (r RedisAnalyticsHandler) PurgeCache() {
	// TODO: Create filename from time parameters
	// TODO: Configurable analytics directory
	// TODO: Configurable cache purge writer (e.g. PG)

	outfile, _ := os.Create("test.analytics.csv")
	defer outfile.Close()
	writer := csv.NewWriter(outfile)

	var handlers = []string{"METHOD", "PATH", "SIZE", "UA", "DAY", "MONTH", "YEAR", "HOUR", "RESPONSE"}

	err := writer.Write(handlers)
	if err != nil {
		log.Error("Failed to write file handlers!")
		log.Error(err)
	} else {
		keyValueMap := r.Store.GetKeyAndValues()
	}
}
