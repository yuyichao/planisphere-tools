package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/go-git/go-git/v5"
)

// encrypt the provided bytes for the provided encryption
// keys recipients. Returns the encrypted content bytes.
// func encrypt(d []byte, encryptionKeys *openpgp.EntityList) ([]byte, error) {
func encrypt(d []byte, encryptionKeys []*openpgp.Entity) ([]byte, error) {
	buffer := &bytes.Buffer{}
	var armoredWriter io.WriteCloser
	var cipheredWriter io.WriteCloser
	var err error

	// Create an openpgp armored cipher writer pointing on our
	// buffer
	armoredWriter, err = armor.Encode(buffer, "PGP MESSAGE", nil)
	if err != nil {
		return nil, errors.New("bad writer")
	}

	// Create an encrypted writer using the provided encryption keys
	cipheredWriter, err = openpgp.Encrypt(armoredWriter, encryptionKeys, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("bad cipher: %x", err)
	}

	// Write (encrypts on the fly) the provided bytes to
	// cipheredWriter
	if _, err := cipheredWriter.Write(d); err != nil {
		return nil, errors.New("bad ciphered writer")
	}

	if err := cipheredWriter.Close(); err != nil {
		slog.Warn("error closing ciphered writer", "error", err)
	}
	if err := armoredWriter.Close(); err != nil {
		slog.Warn("error closing armored writer", "error", err)
	}

	return buffer.Bytes(), nil
}

func readEntity(name string) (*openpgp.Entity, error) {
	f, err := os.Open(path.Clean(name))
	if err != nil {
		return nil, err
	}
	defer dclose(f)
	block, err := armor.Decode(f)
	if err != nil {
		return nil, err
	}
	return openpgp.ReadEntity(packet.NewReader(block.Body))
}

// collectGPGPubKeys returns an EntityList from a given url
// func collectGPGPubKeys(fp string) (*openpgp.EntityList, error) {
func collectGPGPubKeys(fp string) ([]*openpgp.Entity, error) {
	if fp == "" {
		gitlabKeysURL := "https://gitlab.oit.duke.edu/oit-ssi-systems/staff-public-keys.git"
		subDir := "linux"
		tmpdir, err := os.MkdirTemp("", "gpg-pub-tmpdir")
		if err != nil {
			return nil, err
		}
		defer dRemoveAll(tmpdir)
		_, err = git.PlainClone(tmpdir, false, &git.CloneOptions{
			URL:               gitlabKeysURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			return nil, err
		}
		fp = path.Join(tmpdir, subDir)
		slog.Info("using keys from", "subdir", subDir, "url", gitlabKeysURL)
	}

	matches, err := filepath.Glob(fmt.Sprintf("%v/*.gpg", fp))
	if err != nil {
		return nil, err
	}
	els := []*openpgp.Entity{}
	for _, pubKeyFile := range matches {
		e, err := readEntity(pubKeyFile)
		if err != nil {
			slog.Warn("error opening gpg file", "pubkey", pubKeyFile)
			continue
		}
		armoredWriter, err := armor.Encode(&bytes.Buffer{}, "PGP MESSAGE", nil)
		if err != nil {
			return nil, errors.New("bad writer")
		}
		if _, err := openpgp.Encrypt(armoredWriter, []*openpgp.Entity{e}, nil, nil, nil); err != nil {
			slog.Warn("error testing encryption with key", "error", err)
			continue
		}
		els = append(els, e)
	}
	if len(els) == 0 {
		return nil, errors.New("no gpg keys found")
	}
	return els, nil
}

func dclose(c io.Closer) {
	if err := c.Close(); err != nil {
		fmt.Fprint(os.Stderr, "error closing file\n")
	}
}

func dRemoveAll(path string) {
	if err := os.RemoveAll(path); err != nil {
		fmt.Fprintf(os.Stderr, "error removing path: %v\n", path)
	}
}
