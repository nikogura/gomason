package gomason

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadWithBearerToken(t *testing.T) {
	receivedAuth := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := server.Client()

	err := Upload(client, server.URL+"/test/file", strings.NewReader("test content"), "md5", "sha1", "sha256", "", "", "my-bearer-token")
	require.NoError(t, err, "upload with bearer token should succeed")
	assert.Equal(t, "Bearer my-bearer-token", receivedAuth, "bearer token should be sent as Authorization header")
}

func TestUploadWithBasicAuth(t *testing.T) {
	receivedAuth := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := server.Client()

	err := Upload(client, server.URL+"/test/file", strings.NewReader("test content"), "md5", "sha1", "sha256", "user", "pass", "")
	require.NoError(t, err, "upload with basic auth should succeed")
	assert.True(t, strings.HasPrefix(receivedAuth, "Basic "), "basic auth header should be sent when no bearer token")
}

func TestUploadBearerTokenTakesPrecedence(t *testing.T) {
	receivedAuth := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := server.Client()

	// Both basic auth creds and bearer token provided — bearer should win.
	err := Upload(client, server.URL+"/test/file", strings.NewReader("test content"), "md5", "sha1", "sha256", "user", "pass", "my-token")
	require.NoError(t, err, "upload should succeed")
	assert.Equal(t, "Bearer my-token", receivedAuth, "bearer token should take precedence over basic auth")
}

func TestUploadNoAuth(t *testing.T) {
	receivedAuth := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := server.Client()

	// No credentials at all — should still work (some repos don't require auth).
	err := Upload(client, server.URL+"/test/file", strings.NewReader("test content"), "md5", "sha1", "sha256", "", "", "")
	require.NoError(t, err, "upload without auth should succeed")
	// Basic auth with empty user/pass still sends a header, so we just check it's not Bearer.
	assert.False(t, strings.HasPrefix(receivedAuth, "Bearer "), "no bearer token should be sent")
}

func TestGetCredentialsWithBearerToken(t *testing.T) {
	meta := Metadata{
		PublishInfo: PublishInfo{
			BearerToken: "metadata-token",
		},
	}

	g := &Gomason{
		Config: UserConfig{},
	}

	_, _, bearerToken, err := g.GetCredentials(meta)
	require.NoError(t, err, "GetCredentials should not error")
	assert.Equal(t, "metadata-token", bearerToken, "bearer token from metadata should be returned")
}

func TestGetCredentialsWithBearerTokenFunc(t *testing.T) {
	meta := Metadata{
		PublishInfo: PublishInfo{
			BearerTokenFunc: "echo test-token-from-func",
		},
	}

	g := &Gomason{
		Config: UserConfig{},
	}

	_, _, bearerToken, err := g.GetCredentials(meta)
	require.NoError(t, err, "GetCredentials should not error")
	assert.Equal(t, "test-token-from-func", bearerToken, "bearer token from func should be returned")
}

func TestGetCredentialsBearerTokenFuncTakesPrecedence(t *testing.T) {
	meta := Metadata{
		PublishInfo: PublishInfo{
			BearerToken:     "static-token",
			BearerTokenFunc: "echo dynamic-token",
		},
	}

	g := &Gomason{
		Config: UserConfig{},
	}

	_, _, bearerToken, err := g.GetCredentials(meta)
	require.NoError(t, err, "GetCredentials should not error")
	assert.Equal(t, "dynamic-token", bearerToken, "bearer token func should take precedence over static token")
}

func TestGetCredentialsUserConfigBearerTokenOverridesMetadata(t *testing.T) {
	meta := Metadata{
		PublishInfo: PublishInfo{
			BearerToken: "metadata-token",
		},
	}

	g := &Gomason{
		Config: UserConfig{
			User: UserInfo{
				BearerToken: "user-config-token",
			},
		},
	}

	_, _, bearerToken, err := g.GetCredentials(meta)
	require.NoError(t, err, "GetCredentials should not error")
	assert.Equal(t, "user-config-token", bearerToken, "user config bearer token should override metadata")
}

func TestUploadChecksumHeaders(t *testing.T) {
	receivedHeaders := make(http.Header)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := server.Client()

	err := Upload(client, server.URL+"/test/file", strings.NewReader("data"), "test-md5", "test-sha1", "test-sha256", "", "", "token123")
	require.NoError(t, err, "upload should succeed")
	assert.Equal(t, "test-md5", receivedHeaders.Get("X-Checksum-Md5"), "md5 checksum header should be set")
	assert.Equal(t, "test-sha1", receivedHeaders.Get("X-Checksum-Sha1"), "sha1 checksum header should be set")
	assert.Equal(t, "test-sha256", receivedHeaders.Get("X-Checksum-Sha256"), "sha256 checksum header should be set")
}

func TestUploadServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := server.Client()

	err := Upload(client, server.URL+"/test/file", strings.NewReader("data"), "md5", "sha1", "sha256", "", "", "token")
	require.Error(t, err, "upload should fail on server error")
	assert.Contains(t, err.Error(), strconv.Itoa(http.StatusInternalServerError), "error should contain status code")
}

func TestReadMetadataWithBearerToken(t *testing.T) {
	metadataJSON := `{
		"version": "1.0.0",
		"package": "github.com/example/tool",
		"description": "Test tool",
		"publishing": {
			"targets": [],
			"bearertoken": "static-token-in-metadata",
			"bearertokenfunc": "echo dynamic-token"
		}
	}`

	tmpFile := fmt.Sprintf("%s/metadata-bearer.json", TestTmpDir)

	err := writeTestFile(tmpFile, metadataJSON)
	if err != nil {
		t.Fatalf("failed to write test metadata file: %s", err)
	}

	meta, err := ReadMetadata(tmpFile)
	require.NoError(t, err, "ReadMetadata should not error")
	assert.Equal(t, "static-token-in-metadata", meta.PublishInfo.BearerToken, "bearer token should be parsed from metadata")
	assert.Equal(t, "echo dynamic-token", meta.PublishInfo.BearerTokenFunc, "bearer token func should be parsed from metadata")
}

func writeTestFile(path, content string) (err error) {
	err = os.WriteFile(path, []byte(content), 0644)

	return err
}
