package key

import (
	"os"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Load attempts to load private key information from the given private key file.
func Load(privKeyFile string) (crypto.PrivKey, error) {
	log.WithField("privKeyFile", privKeyFile).Infoln("Loading private key file...")

	if _, err := os.Stat(privKeyFile); err != nil {
		return nil, err
	}

	dat, err := os.ReadFile(privKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading key data")
	}

	key, err := crypto.UnmarshalPrivateKey(dat)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal private key")
	}

	return key, nil
}

// Create generates a new 256-bit Ed25519 key pair and save the private key to the given file.
func Create(privKeyFile string) (crypto.PrivKey, error) {
	log.WithField("privKeyFile", privKeyFile).Infoln("Generating new key pair...")

	key, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	if err != nil {
		return nil, errors.Wrap(err, "generate key pair")
	}

	keyDat, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "marshal private key")
	}

	if err = os.WriteFile(privKeyFile, keyDat, 0o644); err != nil {
		return nil, errors.Wrap(err, "write marshaled private key")
	}

	return key, nil
}
