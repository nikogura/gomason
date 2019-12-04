package gomason

func TestMetadataObj() (metadata Metadata) {
	metadata = Metadata{
		Package:     TestModuleName(),
		Version:     "0.1.0",
		Description: "Test Project for Gomason.",
		BuildInfo: BuildInfo{
			PrepCommands: []string{
				"echo \"GOPATH is: ${GOPATH}\"",
			},
			Targets: []BuildTarget{{Name: "linux/amd64"}, {Name: "darwin/amd64"}},
		},
		SignInfo: SignInfo{
			Program: "gpg",
			Email:   "gomason-tester@foo.com",
		},
		PublishInfo: PublishInfo{
			Targets: []PublishTarget{
				{
					Source:      "testproject_darwin_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/darwin/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
				{
					Source:      "testproject_linux_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
			},
			TargetsMap: map[string]PublishTarget{
				"testproject_darwin_amd64": {
					Source:      "testproject_darwin_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/darwin/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
				"testproject_linux_amd64": {
					Source:      "testproject_linux_amd64",
					Destination: "{{.Repository}}/testproject/{{.Version}}/linux/amd64/testproject",
					Signature:   true,
					Checksums:   true,
				},
			},
		},
	}

	return metadata
}

func TestModuleName() string {
	return "github.com/nikogura/testproject"
}
