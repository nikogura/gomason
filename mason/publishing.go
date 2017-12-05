package mason

// FEATURE  publish needs an option to create and upload the foo.md5, foo.sha1, foo.sha256 files if the repo doesn't do it automatically like artifactory does.

// FEATURE publish needs to be able to upload to a target directory or directory pattern

// Feature needs a mechanism to upload 'extra'.  Src, dst.  dst can take some parameters such as arch,  os, version

// PublishBinary publishes the binary to wherever you have it configured to go
func PublishBinary(meta Metadata, cwd string, binary string, binaryPrefix string, osname string, archname string) (err error) {

	// TODO Implement PublishBinary

	return err
}
