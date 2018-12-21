package gomason

func testMetadataObj() (metadata Metadata) {
	metadata = Metadata{
		Package:     testModuleName(),
		Version:     "0.1.0",
		Description: "Test Project for Gomason.",
		BuildInfo: BuildInfo{
			PrepCommands: []string{
				"GO111MODULE=off go get k8s.io/client-go/...",
				"cd ${GOPATH}/src/k8s.io/client-go && git checkout v10.0.0",
				"cd ${GOPATH}/src/k8s.io/client-go && godep restore ./...",
			},
			Targets: []BuildTarget{{Name: "linux/amd64"}, {Name: "darwin/amd64"}},
		},
		SignInfo: SignInfo{
			Program: "gpg",
			Email:   "gomason-tester@foo.com",
		},
		PublishInfo: PublishInfo{
			Targets:    make([]PublishTarget, 0),
			TargetsMap: make(map[string]PublishTarget),
		},
	}

	return metadata
}

func testModuleName() string {
	return "github.com/nikogura/testproject"
}
