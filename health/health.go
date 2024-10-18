package health

import (
	"loadbalancer/serverpool"
	"log"
	"time"
)

func HealthCheck(pool *serverpool.ServerPool, healthCheckIntervalInSeconds int) {
	t := time.NewTicker(time.Second * time.Duration(healthCheckIntervalInSeconds))
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			pool.HealthCheck()
			log.Println("Health check completed")
		}
	}
}
