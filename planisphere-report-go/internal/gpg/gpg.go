package gpg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"
)

// Encrypt the provided bytes for the provided encryption
// keys recipients. Returns the encrypted content bytes.
func Encrypt(d []byte, encryptionKeys *openpgp.EntityList) ([]byte, error) {
	var buffer *bytes.Buffer = &bytes.Buffer{}
	var armoredWriter io.WriteCloser
	var cipheredWriter io.WriteCloser
	var err error

	// Create an openpgp armored cipher writer pointing on our
	// buffer
	armoredWriter, err = armor.Encode(buffer, "PGP MESSAGE", nil)
	if err != nil {
		return nil, errors.New("Bad Writer")
	}

	// Create an encrypted writer using the provided encryption keys
	cipheredWriter, err = openpgp.Encrypt(armoredWriter, *encryptionKeys, nil, nil, nil)
	if err != nil {
		return nil, errors.New("Bad Cipher")
	}

	// Write (encrypts on the fly) the provided bytes to
	// cipheredWriter
	_, err = cipheredWriter.Write(d)
	if err != nil {
		return nil, errors.New("Bad Ciphered Writer")
	}

	cipheredWriter.Close()
	armoredWriter.Close()

	return buffer.Bytes(), nil
}

func ReadEntity(name string) (*openpgp.Entity, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	block, err := armor.Decode(f)
	if err != nil {
		return nil, err
	}
	return openpgp.ReadEntity(packet.NewReader(block.Body))
}

func CollectGPGPubKeys(fp string) (*openpgp.EntityList, error) {
	var els openpgp.EntityList

	if fp == "" {
		gitlabKeysURL := "https://gitlab.oit.duke.edu/oit-ssi-systems/staff-public-keys.git"
		subDir := "linux"
		tmpdir, err := os.MkdirTemp("", "gpg-pub-tmpdir")
		defer os.RemoveAll(tmpdir)
		if err != nil {
			return nil, err
		}
		_, err = git.PlainClone(tmpdir, false, &git.CloneOptions{
			URL:               gitlabKeysURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			return nil, err
		}
		fp = path.Join(tmpdir, subDir)
		log.Info().Str("subdir", subDir).Str("url", gitlabKeysURL).Msg("Using keys from")
	}

	matches, err := filepath.Glob(fmt.Sprintf("%v/*.gpg", fp))
	if err != nil {
		return nil, err
	}
	for _, pubKeyFile := range matches {
		e, err := ReadEntity(pubKeyFile)
		if err != nil {
			log.Warn().Str("pubkey", pubKeyFile).Msg("Error opening gpg file")
			continue
		}
		els = append(els, e)
	}
	if len(els) == 0 {
		return nil, errors.New("No gpg keys found")
	}
	return &els, nil
}
