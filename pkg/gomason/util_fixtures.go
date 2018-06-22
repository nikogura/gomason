package gomason

func testMetaDataJson() string {
	return `{
	"version": "0.1.0",
	"package": "github.com/nikogura/gomason",
	"description": "A tool for building and testing your project in a clean GOPATH.",
	"building": {
		"targets": [
			"linux/amd64"
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
