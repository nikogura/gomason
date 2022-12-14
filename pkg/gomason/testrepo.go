package gomason

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

// TestRepo a fake repository server.  Basically an in-memory http server that can be used as a test fixture for testing the internal API.  Cool huh?
type TestRepo struct{}

// Run runs the test repository server.
func (tr *TestRepo) Run(port int) (err error) {

	logrus.Infof("Running test artifact server on port %d", port)

	http.HandleFunc("/repo/tool/", tr.HandlerTool)

	err = http.ListenAndServe(fmt.Sprintf("localhost:%s", strconv.Itoa(port)), nil)

	return err
}

// HandlerTool handles requests publishing a tool in the test repo
func (tr *TestRepo) HandlerTool(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("*TestRepo: Request for %s*", r.URL.Path)

	// we just return 200.  We're not doing anything beyond providing an endpoint for the client to hit.
	w.WriteHeader(200)
}
