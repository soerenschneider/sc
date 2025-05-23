package format

import (
	"crypto/x509"
	"errors"

	"github.com/soerenschneider/sc/internal/vault"
	"go.uber.org/multierr"
)

type MultiKeyPairStorage struct {
	sinks []*KeyPairStorage
}

func NewMultiKeyPairSink(sinks ...*KeyPairStorage) (*MultiKeyPairStorage, error) {
	if nil == sinks {
		return nil, errors.New("no sinks provided")
	}

	return &MultiKeyPairStorage{sinks: sinks}, nil
}

func (f *MultiKeyPairStorage) WriteCert(certData *vault.PkiCertData) error {
	var err error
	for _, sink := range f.sinks {
		writeErr := sink.WriteCert(certData)
		if writeErr != nil {
			err = multierr.Append(err, writeErr)
		}
	}

	return err
}

func (f *MultiKeyPairStorage) ReadCert() (*x509.Certificate, error) {
	for _, sink := range f.sinks {
		cert, err := sink.ReadCert()
		if err == nil {
			return cert, err
		}
	}

	return nil, errors.New("could not read any cert")
}
