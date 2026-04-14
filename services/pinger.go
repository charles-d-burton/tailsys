package services

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/charles-d-burton/tailsys/models"
)

type Pinger struct {
	DB *sql.DB
	ID string
}

// Ping HTTP service for the service to ping clients and provide response time
func (p *Pinger) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received ping request")
	var in models.PingRequest
	if err := models.ReadYAML(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now()
	latency := now.Sub(in.Ping)

	out := models.PongResponse{
		Ping:           now,
		InboundLatency: float32(latency.Milliseconds()),
	}
	models.WriteYAML(w, http.StatusOK, out)
}
