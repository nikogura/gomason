// Copyright © 2026 Nik Ogura <nik.ogura@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testfixtures

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// TestRepoReadHeaderTimeout bounds how long the test repo will wait on a
// client's request headers.
const TestRepoReadHeaderTimeout = 10 * time.Second

// TestRepo is a fake repository server: an in-memory http server usable as a
// test fixture for exercising the publishing code paths.
type TestRepo struct{}

// Run runs the test repository server on the given port.  It blocks until the
// server stops, so callers generally want it in a goroutine.
func (tr *TestRepo) Run(port int) (err error) {
	logrus.Debugf("Running test artifact server on port %d", port)

	// A dedicated mux rather than http.DefaultServeMux, so that running more
	// than one TestRepo in a process doesn't panic on a duplicate route.
	mux := http.NewServeMux()
	mux.HandleFunc("/repo/tool/", tr.HandlerTool)

	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", strconv.Itoa(port)),
		Handler:           mux,
		ReadHeaderTimeout: TestRepoReadHeaderTimeout,
	}

	err = server.ListenAndServe()

	return err
}

// HandlerTool handles requests publishing a tool in the test repo.
func (tr *TestRepo) HandlerTool(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("*TestRepo: Request for %s*", r.URL.Path)

	// We just return 200.  We're not doing anything beyond providing an
	// endpoint for the client to hit.
	w.WriteHeader(http.StatusOK)
}
