package gomason

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// It's a good default.  You can install it anywhere.
const defaultSigningProgram = "gpg"

// SignBinary  signs the given binary based on the entity and program given in metadata file, possibly overridden by information in ~/.gomason
func (g *Gomason) SignBinary(meta Metadata, binary string) (err error) {
	logrus.Debugf("Preparing to sign file %s", binary)

	// pull signing info out of metadata file
	signInfo := meta.SignInfo
	signProg := signInfo.Program
	if signProg == "" {
		signProg = defaultSigningProgram
	}

	logrus.Debugf("Signing program is %s", signProg)

	signEntity := signInfo.Email

	config := g.Config

	// email from .gomason overrides metadata
	if config.User.Email != "" {
		signEntity = config.User.Email
	}

	// program from .gomason overrides metadata
	if config.Signing.Program != "" {
		signProg = config.Signing.Program
	}

	if signEntity == "" {
		err = fmt.Errorf("Cannot sign without a signing entity (email).\n\nSet 'signing' section in metadata file, or create ~/.gomason with the appropriate content.\n\nSee https://github.com/nikogura/gomason#config-reference for details.\n\n")

		return err
	}

	logrus.Debugf("Signing %s with identity %s.", binary, signEntity)

	switch signProg {
	// insert other signing types here
	default:
		logrus.Debug("Signing with default program.")
		err = SignGPG(binary, signEntity, meta)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to run %q", signProg))
			return err
		}
	}

	return err
}

// VerifyBinary will verify the signature of a signed binary.
func VerifyBinary(binary string, meta Metadata) (ok bool, err error) {
	// pull signing info out of metadata file
	signInfo := meta.SignInfo

	signProg := signInfo.Program
	if signProg == "" {
		signProg = defaultSigningProgram
	}
	switch signProg {
	// insert other signing types here
	default:
		logrus.Debugf("Verifying with default program.")
		ok, err = VerifyGPG(binary, meta)
		if err != nil {
			err = errors.Wrapf(err, "failed to run %q", signProg)
			return ok, err
		}
	}

	return ok, err
}

// SignGPG signs a given binary with GPG using the given signing entity.
func SignGPG(binary string, signingEntity string, meta Metadata) (err error) {
	shellCmd, err := exec.LookPath("gpg")
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("can't find signing program 'gpg' in path.  Is it installed?"))
		return err
	}

	var cmd *exec.Cmd

	if keyring, ok := meta.Options["keyring"]; ok {
		// use a custom keyring for testing
		cmd = exec.Command(shellCmd, "--trustdb", meta.Options["trustdb"].(string), "--no-default-keyring", "--keyring", keyring.(string), "-bau", signingEntity, "--yes", binary)

	} else {
		// gpg -bau <email address>  <file>
		// -b detatch  -a ascii armor -u specify user
		cmd = exec.Command(shellCmd, "-bau", signingEntity, "--yes", binary)
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()

	err = cmd.Start()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to run %q", shellCmd))
	}

	err = cmd.Wait()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("error waiting for  %q to finish", shellCmd))
	}

	return err
}

// VerifyGPG  Verifies signatures with gpg.
func VerifyGPG(binary string, meta Metadata) (ok bool, err error) {
	sigFile := fmt.Sprintf("%s.asc", binary)

	shellCmd, err := exec.LookPath("gpg")
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("can't find signing program 'gpg' in path.  Is it installed?"))
		return ok, err
	}

	var cmd *exec.Cmd

	if keyring, ok := meta.Options["keyring"]; ok {
		// use a custom keyring for testing
		cmd = exec.Command(shellCmd, "--trustdb", meta.Options["trustdb"].(string), "--no-default-keyring", "--keyring", keyring.(string), "--verify", sigFile)

	} else {
		// gpg --verify  <file>
		cmd = exec.Command(shellCmd, "--verify", sigFile)
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		err = errors.Wrapf(err, "error verifying %s", sigFile)
		return ok, err
	}

	ok = true

	return ok, err
}
