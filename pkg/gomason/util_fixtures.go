package gomason

func testMetaDataJson() string {
	return `{
	"version": "0.1.0",
	"package": "github.com/nikogura/testproject",
	"description": "Test Project for Gomason.",
	"building": {
		"prepcommands": [
      "echo \"GOPATH is: ${GOPATH}\""
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
	},
	"publishing": {
		"targets": [
			{
				"src": "testproject_darwin_amd64",
				"dst": "{{.Repository}}/testproject/{{.Version}}/darwin/amd64/testproject",
				"sig": true,
				"checksums": true
			},
			{
				"src": "testproject_linux_amd64",
				"dst": "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
				"sig": true,
				"checksums": true
			}
		]
	}
}`
}

func testMetadataFileName() string {
	return "metadata.json"
}
