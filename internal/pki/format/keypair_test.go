package format

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/soerenschneider/sc/internal/pki"
	"github.com/soerenschneider/sc/internal/vault"
)

const (
	privKey = "--- START PRIVATE KEY ---\nSECRETSECRET\n--- END PRIVATE KEY ---"
	cert    = "--- START CERT ---\nCERTCERTCERT\n--- END CERT ---"
	ca      = "--- START CA ---\nCACACACACA\n--- END CA ---"
)

type BufferPod struct {
	Data []byte
}

func (b *BufferPod) Read() ([]byte, error) {
	if len(b.Data) > 0 {
		return b.Data, nil
	}
	return nil, pki.ErrNoCertFound
}

func (b *BufferPod) CanRead() error {
	if len(b.Data) > 0 {
		return nil
	}

	return pki.ErrNoCertFound
}

func (b *BufferPod) Write(data []byte) error {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	b.Data = data
	return nil
}

func (b *BufferPod) CanWrite() error {
	return nil
}

func TestKeyPairSink_WriteCert(t *testing.T) {
	type fields struct {
		ca         StorageImplementation
		cert       StorageImplementation
		privateKey StorageImplementation
	}
	tests := []struct {
		name     string
		fields   fields
		certData *vault.PkiCertData
		wantErr  bool
		wantData string
	}{
		{
			name: "write ca, cert and key to single file",
			certData: &vault.PkiCertData{
				PrivateKey:  []byte(privKey),
				Certificate: []byte(cert),
				CaData:      []byte(ca),
				Csr:         nil,
			},
			fields: fields{
				ca:         nil,
				cert:       nil,
				privateKey: &BufferPod{},
			},
			wantErr:  false,
			wantData: fmt.Sprintf("%s\n%s\n%s\n", cert, ca, privKey),
		},
		{
			name: "write cert and key to single file",
			certData: &vault.PkiCertData{
				PrivateKey:  []byte(privKey),
				Certificate: []byte(cert),
				CaData:      nil,
				Csr:         nil,
			},
			fields: fields{
				ca:         nil,
				cert:       nil,
				privateKey: &BufferPod{},
			},
			wantErr:  false,
			wantData: fmt.Sprintf("%s\n%s\n", cert, privKey),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &KeyPairStorage{
				ca:         tt.fields.ca,
				cert:       tt.fields.cert,
				privateKey: tt.fields.privateKey,
			}
			if err := f.WriteCert(tt.certData); (err != nil) != tt.wantErr {
				t.Errorf("WriteCert() error = %v, wantErr %v", err, tt.wantErr)
			}
			read, err := tt.fields.privateKey.Read()
			if err != nil {
				t.Errorf("Error reading b")
			}

			if !reflect.DeepEqual(string(read), tt.wantData) {
				t.Errorf("KeyPairSinkFromConfig() got = %v, want %v", string(read), tt.wantData)
			}
		})
	}
}

func TestKeyPairSink_writeToIndividualSlots(t *testing.T) {
	type fields struct {
		ca         StorageImplementation
		cert       StorageImplementation
		privateKey StorageImplementation
	}
	type args struct {
		certData *vault.PkiCertData
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		wantCa         []byte
		wantPrivateKey []byte
		wantCert       []byte
	}{
		{
			name: "no newlines",
			fields: fields{
				cert:       &BufferPod{},
				privateKey: &BufferPod{},
			},
			args: args{certData: &vault.PkiCertData{
				PrivateKey:  []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAnVHfwoKsUG1GDVyWB1AFroaKl2ImMBO8EnvGLRrmobIkQvh+\n...\nQN351pgTphi6nlCkGPzkDuwvtxSxiCWXQcaxrHAL7MiJpPzkIBq1\n-----END RSA PRIVATE KEY-----\n"),
				Certificate: []byte("-----BEGIN CERTIFICATE-----\nMIIDzDCCAragAwIBAgIUOd0ukLcjH43TfTHFG9qE0FtlMVgwCwYJKoZIhvcNAQEL\n...\numkqeYeO30g1uYvDuWLXVA==\n-----END CERTIFICATE-----\n"),
				CaData:      []byte("-----BEGIN CERTIFICATE-----\nMIIDUTCCAjmgAwIBAgIJAKM+z4MSfw2mMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNV\n...\nG/7g4koczXLoUM3OQXd5Aq2cs4SS1vODrYmgbioFsQ3eDHd1fg==\n-----END CERTIFICATE-----\n"),
				Csr:         nil,
			}},
			wantErr: false,
			wantPrivateKey: []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAnVHfwoKsUG1GDVyWB1AFroaKl2ImMBO8EnvGLRrmobIkQvh+
...
QN351pgTphi6nlCkGPzkDuwvtxSxiCWXQcaxrHAL7MiJpPzkIBq1
-----END RSA PRIVATE KEY-----
`),
			wantCert: []byte(`-----BEGIN CERTIFICATE-----
MIIDzDCCAragAwIBAgIUOd0ukLcjH43TfTHFG9qE0FtlMVgwCwYJKoZIhvcNAQEL
...
umkqeYeO30g1uYvDuWLXVA==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDUTCCAjmgAwIBAgIJAKM+z4MSfw2mMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNV
...
G/7g4koczXLoUM3OQXd5Aq2cs4SS1vODrYmgbioFsQ3eDHd1fg==
-----END CERTIFICATE-----
`),
			wantCa: []byte(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &KeyPairStorage{
				ca:         tt.fields.ca,
				cert:       tt.fields.cert,
				privateKey: tt.fields.privateKey,
			}
			if err := f.writeToIndividualSlots(tt.args.certData); (err != nil) != tt.wantErr {
				t.Errorf("writeToIndividualSlots() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.fields.privateKey != nil {
				pkData, _ := tt.fields.privateKey.Read()
				if !reflect.DeepEqual(tt.wantPrivateKey, pkData) {
					t.Errorf("wanted pkData=%s, got=%s", string(tt.wantPrivateKey), pkData)
				}
			}

			if tt.fields.cert != nil {
				certData, _ := tt.fields.cert.Read()
				if !reflect.DeepEqual(tt.wantCert, certData) {
					t.Errorf("wanted certData=%s, got=%s", string(tt.wantCert), certData)
				}
			}

			if tt.fields.ca != nil {
				caData, _ := tt.fields.ca.Read()
				if !reflect.DeepEqual(tt.wantCa, caData) {
					t.Errorf("wanted caData=%s, got=%s", string(tt.wantCa), caData)
				}
			}
		})
	}
}

func TestKeyPairSink_writeToPrivateSlot(t *testing.T) {
	type fields struct {
		ca         StorageImplementation
		cert       StorageImplementation
		privateKey StorageImplementation
	}
	type args struct {
		certData *vault.PkiCertData
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		wantCa         []byte
		wantPrivateKey []byte
		wantCert       []byte
	}{
		{
			name: "no newlines",
			fields: fields{
				privateKey: &BufferPod{},
			},
			args: args{certData: &vault.PkiCertData{
				PrivateKey:  []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAnVHfwoKsUG1GDVyWB1AFroaKl2ImMBO8EnvGLRrmobIkQvh+\n...\nQN351pgTphi6nlCkGPzkDuwvtxSxiCWXQcaxrHAL7MiJpPzkIBq1\n-----END RSA PRIVATE KEY-----\n"),
				Certificate: []byte("-----BEGIN CERTIFICATE-----\nMIIDzDCCAragAwIBAgIUOd0ukLcjH43TfTHFG9qE0FtlMVgwCwYJKoZIhvcNAQEL\n...\numkqeYeO30g1uYvDuWLXVA==\n-----END CERTIFICATE-----\n"),
				CaData:      []byte("-----BEGIN CERTIFICATE-----\nMIIDUTCCAjmgAwIBAgIJAKM+z4MSfw2mMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNV\n...\nG/7g4koczXLoUM3OQXd5Aq2cs4SS1vODrYmgbioFsQ3eDHd1fg==\n-----END CERTIFICATE-----\n"),
				Csr:         nil,
			}},
			wantErr: false,
			wantCa:  []byte(""),
			wantPrivateKey: []byte(`-----BEGIN CERTIFICATE-----
MIIDzDCCAragAwIBAgIUOd0ukLcjH43TfTHFG9qE0FtlMVgwCwYJKoZIhvcNAQEL
...
umkqeYeO30g1uYvDuWLXVA==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDUTCCAjmgAwIBAgIJAKM+z4MSfw2mMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNV
...
G/7g4koczXLoUM3OQXd5Aq2cs4SS1vODrYmgbioFsQ3eDHd1fg==
-----END CERTIFICATE-----
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAnVHfwoKsUG1GDVyWB1AFroaKl2ImMBO8EnvGLRrmobIkQvh+
...
QN351pgTphi6nlCkGPzkDuwvtxSxiCWXQcaxrHAL7MiJpPzkIBq1
-----END RSA PRIVATE KEY-----
`),
			wantCert: []byte(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &KeyPairStorage{
				ca:         tt.fields.ca,
				cert:       tt.fields.cert,
				privateKey: tt.fields.privateKey,
			}
			if err := f.writeToPrivateSlot(tt.args.certData); (err != nil) != tt.wantErr {
				t.Errorf("writeToPrivateSlot() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.fields.privateKey != nil {
				pkData, _ := tt.fields.privateKey.Read()
				if !reflect.DeepEqual(tt.wantPrivateKey, pkData) {
					t.Errorf("wanted pkData=%s, got=%s", string(tt.wantPrivateKey), pkData)
				}
			}

			if tt.fields.cert != nil {
				certData, _ := tt.fields.cert.Read()
				if !reflect.DeepEqual(tt.wantCert, certData) {
					t.Errorf("wanted certData=%s, got=%s", string(tt.wantCert), certData)
				}
			}

			if tt.fields.ca != nil {
				caData, _ := tt.fields.ca.Read()
				if !reflect.DeepEqual(tt.wantCa, caData) {
					t.Errorf("wanted caData=%s, got=%s", string(tt.wantCa), caData)
				}
			}
		})
	}
}

func TestKeyPairSink_writeToCertAndCaSlot(t *testing.T) {
	type fields struct {
		ca         StorageImplementation
		cert       StorageImplementation
		privateKey StorageImplementation
	}
	type args struct {
		certData *vault.PkiCertData
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		wantCa         []byte
		wantPrivateKey []byte
		wantCert       []byte
	}{
		{
			name: "no newlines",
			fields: fields{
				privateKey: &BufferPod{},
				ca:         &BufferPod{},
			},
			args: args{certData: &vault.PkiCertData{
				PrivateKey:  []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAnVHfwoKsUG1GDVyWB1AFroaKl2ImMBO8EnvGLRrmobIkQvh+\n...\nQN351pgTphi6nlCkGPzkDuwvtxSxiCWXQcaxrHAL7MiJpPzkIBq1\n-----END RSA PRIVATE KEY-----\n"),
				Certificate: []byte("-----BEGIN CERTIFICATE-----\nMIIDzDCCAragAwIBAgIUOd0ukLcjH43TfTHFG9qE0FtlMVgwCwYJKoZIhvcNAQEL\n...\numkqeYeO30g1uYvDuWLXVA==\n-----END CERTIFICATE-----\n"),
				CaData:      []byte("-----BEGIN CERTIFICATE-----\nMIIDUTCCAjmgAwIBAgIJAKM+z4MSfw2mMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNV\n...\nG/7g4koczXLoUM3OQXd5Aq2cs4SS1vODrYmgbioFsQ3eDHd1fg==\n-----END CERTIFICATE-----\n"),
				Csr:         nil,
			}},
			wantErr: false,
			wantCa: []byte(`-----BEGIN CERTIFICATE-----
MIIDUTCCAjmgAwIBAgIJAKM+z4MSfw2mMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNV
...
G/7g4koczXLoUM3OQXd5Aq2cs4SS1vODrYmgbioFsQ3eDHd1fg==
-----END CERTIFICATE-----
`),
			wantPrivateKey: []byte(`-----BEGIN CERTIFICATE-----
MIIDzDCCAragAwIBAgIUOd0ukLcjH43TfTHFG9qE0FtlMVgwCwYJKoZIhvcNAQEL
...
umkqeYeO30g1uYvDuWLXVA==
-----END CERTIFICATE-----
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAnVHfwoKsUG1GDVyWB1AFroaKl2ImMBO8EnvGLRrmobIkQvh+
...
QN351pgTphi6nlCkGPzkDuwvtxSxiCWXQcaxrHAL7MiJpPzkIBq1
-----END RSA PRIVATE KEY-----
`),
			wantCert: []byte(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &KeyPairStorage{
				ca:         tt.fields.ca,
				cert:       tt.fields.cert,
				privateKey: tt.fields.privateKey,
			}
			if err := f.writeToCertAndCaSlot(tt.args.certData); (err != nil) != tt.wantErr {
				t.Errorf("writeToCertAndCaSlot() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.fields.privateKey != nil {
				pkData, _ := tt.fields.privateKey.Read()
				if !reflect.DeepEqual(tt.wantPrivateKey, pkData) {
					t.Errorf("wanted pkData=%s, got=%s", string(tt.wantPrivateKey), pkData)
				}
			}

			if tt.fields.cert != nil {
				certData, _ := tt.fields.cert.Read()
				if !reflect.DeepEqual(tt.wantCert, certData) {
					t.Errorf("wanted certData=%s, got=%s", string(tt.wantCert), certData)
				}
			}

			if tt.fields.ca != nil {
				caData, _ := tt.fields.ca.Read()
				if !reflect.DeepEqual(tt.wantCa, caData) {
					t.Errorf("wanted caData=%s, got=%s", string(tt.wantCa), caData)
				}
			}
		})
	}
}
