package mason

// Metadata type to represent the metadata.json file
type Metadata struct {
	Version      string   `json:"version"`
	Package      string   `json:"package"`
	Description  string   `json:"description"`
	Path         string   `json:"-"`
	WorkDir      string   `json:"-"`
	ConfigDir    string   `json:"-"`
	SigningKey   string   `json:"-"`
	CodePath     string   `json:"-"`
	GitPath      string   `json:"-"`
	BuildTargets []string `json:"buildtargets,omitempty"`
}
