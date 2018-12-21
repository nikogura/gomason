package gomason

func testMetaDataJson() string {
	return `{
	"version": "0.1.0",
	"package": "github.com/nikogura/testproject",
	"description": "Test Project for Gomason.",
	"building": {
		"prepcommands": [
			"GO111MODULE=off go get k8s.io/client-go/...",
			"cd ${GOPATH}/src/k8s.io/client-go && git checkout v10.0.0",
			"cd ${GOPATH}/src/k8s.io/client-go && godep restore ./..."
		],
		"targets": [
			{
				"name": "linux/amd64"
			},
			{
				"name": "darwin/amd64"
			}
		]
	},
	"signing": {
		"program": "gpg",
		"email": "gomason-tester@foo.com"
	}
}`
}

func testMetadataFileName() string {
	return "metadata.json"
}
