package format

import (
	"bytes"
	"crypto/x509"
	"errors"
	"regexp"

	backend "github.com/soerenschneider/sc/internal/pki"
	"github.com/soerenschneider/sc/internal/vault"
)

// KeyPairStorage offers an interface to read/write keypair data (certificate and private key) and optional ca data.
type KeyPairStorage struct {
	ca         StorageImplementation
	cert       StorageImplementation
	privateKey StorageImplementation
}

type StorageImplementation interface {
	Read() ([]byte, error)
	CanRead() error
	Write([]byte) error
	CanWrite() error
}

func NewKeyPairStorage(cert, privateKey, chain StorageImplementation) (*KeyPairStorage, error) {
	if nil == privateKey {
		return nil, errors.New("empty private key storage provided")
	}

	return &KeyPairStorage{cert: cert, privateKey: privateKey, ca: chain}, nil
}

func (f *KeyPairStorage) ReadCert() (*x509.Certificate, error) {
	var source StorageImplementation
	if f.cert != nil {
		source = f.cert
	} else {
		source = f.privateKey
	}

	data, err := source.Read()
	if err != nil {
		return nil, err
	}

	return backend.ParseCertPem(data)
}

func (f *KeyPairStorage) WriteCert(certData *vault.PkiCertData) error {
	if nil == certData {
		return errors.New("got nil as certData")
	}

	// case 1: write cert, ca and private key to same storage
	if f.cert == nil && f.ca == nil {
		return f.writeToPrivateSlot(certData)
	}

	// case 2: write cert and private to a same storage, write ca (if existent) to dedicated storage
	if f.cert == nil && f.ca != nil {
		return f.writeToCertAndCaSlot(certData)
	}

	// case 3: write to individual storage
	return f.writeToIndividualSlots(certData)
}

func endsWithNewline(data []byte) bool {
	return bytes.HasSuffix(data, []byte("\n"))
}

func (f *KeyPairStorage) writeToPrivateSlot(certData *vault.PkiCertData) error {
	var data = certData.Certificate
	if !endsWithNewline(data) {
		data = append(data, "\n"...)
	}

	if certData.HasCaData() {
		data = append(data, certData.CaData...)
		if !endsWithNewline(data) {
			data = append(data, "\n"...)
		}
	}

	data = append(data, certData.PrivateKey...)
	return f.privateKey.Write(data)
}

func (f *KeyPairStorage) writeToCertAndCaSlot(certData *vault.PkiCertData) error {
	var data = certData.Certificate
	if !endsWithNewline(data) {
		data = append(data, "\n"...)
	}

	data = append(data, certData.PrivateKey...)
	if !endsWithNewline(data) {
		data = append(data, "\n"...)
	}

	if err := f.privateKey.Write(data); err != nil {
		return err
	}

	if certData.HasCaData() {
		caData := certData.CaData
		if !endsWithNewline(caData) {
			caData = append(caData, "\n"...)
		}
		return f.ca.Write(caData)
	}

	return nil
}

var lineBreaksRegex = regexp.MustCompile(`(\r\n?|\n){2,}`)

func fixLineBreaks(input []byte) (ret []byte) {
	ret = []byte(lineBreaksRegex.ReplaceAll(input, []byte("$1")))
	return
}

func (f *KeyPairStorage) writeToIndividualSlots(certData *vault.PkiCertData) error {
	var certRaw = certData.Certificate
	if certData.HasCaData() && f.ca == nil {
		if !endsWithNewline(certRaw) {
			certRaw = append(certRaw, "\n"...)
		}

		certRaw = append(certRaw, certData.CaData...)
	}

	if err := f.cert.Write(certRaw); err != nil {
		return err
	}

	if certData.HasCaData() && f.ca != nil {
		if err := f.ca.Write(certData.CaData); err != nil {
			return err
		}
	}

	if certData.HasPrivateKey() {
		return f.privateKey.Write(fixLineBreaks(certData.PrivateKey))
	}

	return nil
}
