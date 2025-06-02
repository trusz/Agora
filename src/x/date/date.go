package date

import (
	"agora/src/log"
	"time"
)

func FormatDate(date string) string {
	// 2025-06-01T20:10:16Z
	const incomingFormat = "2006-01-02T15:04:05Z"
	const outgoingFormat = "02.01.2006 15:04"

	// Parse the incoming date
	t, err := time.Parse(incomingFormat, date)
	if err != nil {
		log.Error.Printf("Failed to parse date: %v\n", err)
		return date
	}

	// Format the date to the outgoing format
	return t.Format(outgoingFormat)
}
