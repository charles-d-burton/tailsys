package coordination

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/charles-d-burton/tailsys/data/queries"
	"github.com/charles-d-burton/tailsys/models"
	"gopkg.in/yaml.v3"
)

// RegistrationServer struct to contain info for HTTP
type RegistrationServer struct {
	DevMode  bool
	ID       string
	Hostname string
	DB       *sql.DB
}

func (rs *RegistrationServer) createRegistration(nrr *models.NodeRegistrationRequest) error {

	data, err := yaml.Marshal(nrr)
	if err != nil {
		return err
	}

	clientName := nrr.Info.Hostname

	nhr := &queries.RegisteredHostsData{
		Hostname: nrr.Info.Hostname,
		Key:      nrr.Key.Key,
		Data:     data,
	}
	// clientKey := nrr.Key.GetKey()
	fmt.Printf("registering %s\n", clientName)
	err = queries.InsertHostRegistration(rs.DB, nhr)

	if err != nil {
		fmt.Println("could not create registration in database")
		return err
	}
	return nil
}

// Register registers a node with the database when a node sends a request. Returns the server id so the node can verify further requests
func (rs *RegistrationServer) Register(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received coordination registration request")
	var in models.NodeRegistrationRequest
	if err := models.ReadYAML(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(&in)
	in.Accepted = rs.DevMode
	if rs.DevMode {
		fmt.Println("running in dev mode, accepting all incoming connections")
	}
	if err := rs.createRegistration(&in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := models.NodeRegistrationResponse{
		Accepted: rs.DevMode,
		Key:      models.Key{Key: rs.ID},
		Hostname: rs.Hostname,
	}
	models.WriteYAML(w, http.StatusOK, out)
}
