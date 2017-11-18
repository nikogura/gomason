package mason

// Type to represent the metadata.json file
type Metadata struct {
	Path        string
	Package     string
	Version     string
	WorkDir     string
	ConfigDir   string
	SigningKey  string
	CodePath    string
	GitPath     string
	Description string
}
