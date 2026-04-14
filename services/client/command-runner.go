package client

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/charles-d-burton/tailsys/models"
)

// CommandServer struct to contain command runner http
type CommandServer struct {
}

// Command registers the HTTP call and implements behavior for command runner
func (c *CommandServer) Command(w http.ResponseWriter, r *http.Request) {
	var in models.CommandRequest
	if err := models.ReadYAML(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("running command: %s\n", in.Command)
	cnds := strings.Fields(in.Command)
	if len(cnds) == 0 {
		http.Error(w, "empty command", http.StatusBadRequest)
		return
	}
	cmdo := exec.Command(cnds[0], cnds[1:]...)
	out, err := cmdo.CombinedOutput() // Use CombinedOutput to get stderr as well
	fmt.Println(string(out))

	res := models.CommandResponse{
		Timestamp:  time.Now(),
		Successful: true,
		Output:     out,
		ExitCode:   0,
	}

	if err != nil {
		res.Successful = false
		res.ExitCode = 1
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = int32(ee.ExitCode())
		}
	}

	models.WriteYAML(w, http.StatusOK, res)
}
