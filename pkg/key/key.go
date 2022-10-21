package key

import (
	"bufio"
	"encoding/base64"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Load attempts to load private key information from the given private key file.
func Load(privKeyFile string) ([]crypto.PrivKey, error) {
	log.WithField("privKeyFile", privKeyFile).Infoln("Loading private key file...")

	if _, err := os.Stat(privKeyFile); err != nil {
		return nil, err
	}

	file, err := os.Open(privKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "open key data")
	}
	defer file.Close()

	keys := []crypto.PrivKey{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keyStr := scanner.Text()
		if keyStr == "" {
			continue
		}

		keyDat, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, errors.Wrap(err, "decode base64 string")
		}

		key, err := crypto.UnmarshalPrivateKey(keyDat)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal private key")
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// Add generates "count" new 256-bit Ed25519 key pairs and saves the private keys to the given file ( base64 encoded).
func Add(privKeyFile string, count int) ([]crypto.PrivKey, error) {
	log.WithField("privKeyFile", privKeyFile).Infof("Generating %d new key pairs...", count)

	file, err := os.OpenFile(privKeyFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, errors.Wrap(err, "open file")
	}
	defer file.Close()

	keys := []crypto.PrivKey{}
	for i := 0; i < count; i++ {

		log.Infoln("Generating new key pair...")

		key, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
		if err != nil {
			return nil, errors.Wrap(err, "generate key pair")
		}

		keyDat, err := crypto.MarshalPrivateKey(key)
		if err != nil {
			return nil, errors.Wrap(err, "marshal private key")
		}

		if _, err = file.WriteString(base64.StdEncoding.EncodeToString(keyDat) + "\n"); err != nil {
			return nil, errors.Wrap(err, "write private key")
		}

		keys = append(keys, key)
	}

	return keys, nil
}
