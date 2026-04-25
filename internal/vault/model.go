package vault

import "strings"

type SshSignatureRequest struct {
	PublicKey  string
	Ttl        string
	Principals []string
	Extensions map[string]string

	VaultRole string
}
type PkiSignatureArgs struct {
	CommonName string
	Ttl        string
	IpSans     []string
	AltNames   []string
}

type PkiIssueArgs struct {
	CommonName string
	Ttl        string
	IpSans     []string
	AltNames   []string
	Role       string
}

type PkiCertData struct {
	PrivateKey  []byte //#nosec:G117
	Certificate []byte
	CaData      []byte
	Csr         []byte
}

func (cert *PkiCertData) AsContainer() string {
	var buffer strings.Builder

	if cert.HasCaData() {
		buffer.Write(cert.CaData)
		buffer.Write([]byte("\n"))
	}

	buffer.Write(cert.Certificate)
	buffer.Write([]byte("\n"))

	if cert.HasPrivateKey() {
		buffer.Write(cert.PrivateKey)
		buffer.Write([]byte("\n"))
	}

	return buffer.String()
}

func (cert *PkiCertData) HasPrivateKey() bool {
	return len(cert.PrivateKey) > 0
}

func (cert *PkiCertData) HasCertificate() bool {
	return len(cert.Certificate) > 0
}

func (cert *PkiCertData) HasCaData() bool {
	return len(cert.CaData) > 0
}

type PkiSignature struct {
	Certificate []byte
	CaData      []byte
	Serial      string
}

func (cert *PkiSignature) HasCaData() bool {
	return len(cert.CaData) > 0
}

type Kv2SyncConfig struct {
	SecretPath    string                 `yaml:"secret_path" validate:"required"`
	DestUri       string                 `yaml:"dest_uri" validate:"required,filepath"`
	Formatter     string                 `yaml:"formatter" validate:"oneof=env template json yaml"`
	FormatterArgs map[string]interface{} `yaml:"formatter_args"`
}

type StorageImplementation interface {
	Read() ([]byte, error)
	CanRead() error
	Write([]byte) error
}
