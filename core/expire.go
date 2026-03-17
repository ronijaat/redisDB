package core

import (
	"log"
	"time"
)

func expireSample() float32 {
	var limit int = 20
	var expiredCount int = 0

	for key, val := range store {
		if val.ExpiresAt != -1 {
			limit--

			if val.ExpiresAt <= time.Now().UnixMilli() {
				delete(store, key)
				expiredCount++
			}
		}

		if limit == 0 {
			break
		}
	}
	return float32(expiredCount) / float32(limit)
}

func DeleteExpiredKeys() {
	for {
		frac := expireSample()

		if frac <= 0.25 {
			break
		}
	}
	log.Println("deleted the expired but undeleted keys. total keys", len(store))
}
