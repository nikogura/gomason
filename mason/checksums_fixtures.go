package mason

import "fmt"

func testFileContent() string {
	return `the quick fox jumped over the lazy brown dog`
}

func testFileMd5() string {
	return "356b5768c6964531f678781446840b76"
}

func testFileSha1() string {
	return "041b2390cd9697ba6b9f57b532b0aa5ac183736b"
}

func testFileSha256() string {
	return "e088f8b9456b8a91a48159497ac425a4c3cdcad3ad81cc3a269618209dee033b"
}

func testRawUrl() string {
	return "http://localhost:8081/artifactory/repo-local/foo/{{.Version}}/linux/amd64/foo"
}

func testParsedUrl(version string) string {
	return fmt.Sprintf("http://localhost:8081/artifactory/repo-local/foo/%s/linux/amd64/foo", version)
}

func testAllChecksums() []string {
	return []string{testFileMd5(), testFileSha1(), testFileSha256()}
}
