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
	}
}`
}

func testMetadataFileName() string {
	return "metadata.json"
}
