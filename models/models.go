package models

import (
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
)

// WriteYAML writes a YAML response
func WriteYAML(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(status)
	yaml.NewEncoder(w).Encode(data)
}

// ReadYAML reads a YAML request body
func ReadYAML(r *http.Request, data interface{}) error {
	return yaml.NewDecoder(r.Body).Decode(data)
}

// OSType enum equivalent
type OSType int32

const (
	OS_TYPE_UNSPECIFIED OSType = 0
	LINUX               OSType = 1
	MACOS               OSType = 2
	WINDOWS             OSType = 3
)

// SystemType enum equivalent
type SystemType int32

const (
	COORDINATION_SERVER SystemType = 0
	CLIENT              SystemType = 1
	CLI                 SystemType = 2
)

// Key model
type Key struct {
	Key string `yaml:"key"`
}

// SysInfo model
type SysInfo struct {
	Hostname string    `yaml:"hostname"`
	Type     OSType    `yaml:"type"`
	IP       string    `yaml:"ip"`
	LastSeen time.Time `yaml:"lastSeen"`
	Port     string    `yaml:"port"`
}

// NodeRegistrationRequest model
type NodeRegistrationRequest struct {
	Info       SysInfo    `yaml:"info"`
	Key        Key        `yaml:"key"`
	SystemType SystemType `yaml:"systemType"`
	Accepted   bool       `yaml:"accepted"`
	TLSKey     string     `yaml:"tlskey"`
	TLSCert    string     `yaml:"tlscert"`
}

// NodeRegistrationResponse model
type NodeRegistrationResponse struct {
	Accepted bool   `yaml:"accepted"`
	Key      Key    `yaml:"key"`
	Hostname string `yaml:"hostname"`
}

// PingRequest model
type PingRequest struct {
	Ping time.Time `yaml:"ping"`
}

// PongResponse model
type PongResponse struct {
	Ping           time.Time `yaml:"ping"`
	InboundLatency float32   `yaml:"inboundLatency"`
	Key            Key       `yaml:"key"`
}

// CommandRequest model
type CommandRequest struct {
	Requested time.Time `yaml:"requested"`
	Command   string    `yaml:"command"`
	Key       Key       `yaml:"key"`
}

// CommandResponse model
type CommandResponse struct {
	Timestamp  time.Time `yaml:"timestamp"`
	Successful bool      `yaml:"successful"`
	Output     []byte    `yaml:"output"`
	ExitCode   int32     `yaml:"exitCode"`
}

// NodeQuery model
type NodeQuery struct {
	Pattern string `yaml:"pattern"`
	Key     Key    `yaml:"key"`
}

// NodeQueryResponse model
type NodeQueryResponse struct {
	Nodes []string `yaml:"nodes"`
}

// AggregateResponses model
type AggregateResponses struct {
	Response []CommandResponse `yaml:"response"`
}

// CommanderRequest model
type CommanderRequest struct {
	Pattern string `yaml:"pattern"`
	Command string `yaml:"command"`
}
