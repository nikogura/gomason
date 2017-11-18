package mason

func testMetadataObj() (metadata Metadata) {
	metadata = Metadata{
		Package:     "github.com/nikogura/gomason",
		Version:     "0.1.0",
		Description: "A tool for building and testing your project in a clean GOPATH.",
	}

	return metadata
}

func testMetaDataJson() string {
	return `{
	"version": "0.1.0",
	"package": "github.com/nikogura/gomason",
	"description": "A tool for building and testing your project in a clean GOPATH."
}`
}

func testMetadataFileName() string {
	return "metadata.json"
}
