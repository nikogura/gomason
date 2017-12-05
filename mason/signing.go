package mason

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"os/exec"
	"os/user"
)

// It's a good default.  You can install it anywhere.
const defaultSigningProgram = "gpg"

// SignBinary  signs the given binary based on the entity and program given in metadata.json, possibly overridden by information in ~/.gomason
func SignBinary(meta Metadata, binary string, verbose bool) (err error) {

	// pull signing info out of metadata.json
	signInfo := meta.SignInfo

	signProg := signInfo.Program
	if signProg == "" {
		signProg = defaultSigningProgram
	}

	signEntity := signInfo.Email

	// pull per-user signing info out of ~/.gomason if present
	userObj, err := user.Current()
	if err != nil {
		err = fmt.Errorf("failed to get current user: %s", err)
		return err
	}

	homeDir := userObj.HomeDir

	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		err = fmt.Errorf("user %s's homedir %s does not exist", userObj.Name, homeDir)
		return err
	}

	perUserConfigFile := fmt.Sprintf("%s/.gomason", homeDir)

	if _, err := os.Stat(perUserConfigFile); !os.IsNotExist(err) {
		cfg, err := ini.Load(perUserConfigFile)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to load per user config file %s", perUserConfigFile))
			return err
		}

		userSection, _ := cfg.GetSection("user")
		if userSection != nil {
			if userSection.HasKey("email") {
				key, _ := userSection.GetKey("email")
				signEntity = key.Value()
			}
		}

		signingSection, _ := cfg.GetSection("signing")
		if signingSection != nil {
			if signingSection.HasKey("program") {
				key, _ := signingSection.GetKey("program")
				signProg = key.Value()
			}
		}
	}

	if signEntity == "" {
		err = fmt.Errorf("Cannot sign without a signing entity (email).\n\nSet 'signing' section in metadata.json, or create ~/.gomason with the appropriate content.\n\nSee https://github.com/nikogura/gomason#config-reference for details.\n\n")

		return err
	}

	if verbose {
		log.Printf("Signing %s with identity %s.", binary, signEntity)
	}

	switch signProg {
	// insert other signing types here
	default:
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
	// pull signing info out of metadata.json
	signInfo := meta.SignInfo

	signProg := signInfo.Program
	if signProg == "" {
		signProg = defaultSigningProgram
	}
	switch signProg {
	// insert other signing types here
	default:
		ok, err = VerifyGPG(binary, meta)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to run %q", signProg))
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
		cmd = exec.Command(shellCmd, "--trustdb", meta.Options["trustdb"].(string), "--no-default-keyring", "--keyring", keyring.(string), "-bau", signingEntity, binary)

	} else {
		// gpg -bau <email address>  <file>
		// -b detatch  -a ascii armor -u specify user
		cmd = exec.Command(shellCmd, "-bau", signingEntity, binary)
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to run %q", shellCmd))
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
		log.Printf("Verification Error: %s", err)
		return ok, err
	}

	ok = true

	return ok, err
}
