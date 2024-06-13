package health

import (
	"loadbalancer/serverpool"
	"log"
	"time"
)

func HealthCheck(pool *serverpool.ServerPool) {
	t := time.NewTicker(2 * time.Minute)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			pool.HealthCheck()
			log.Println("Health check completed")
		}
	}
}
