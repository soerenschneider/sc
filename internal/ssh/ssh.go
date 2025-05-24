package vault

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type CertInfo struct {
	Type        string
	Serial      uint64
	ValidAfter  time.Time
	ValidBefore time.Time
}

func (l *CertInfo) GetPercentage() float32 {
	total := l.ValidBefore.Sub(l.ValidAfter).Seconds()
	if total == 0 {
		return 0.
	}

	left := time.Until(l.ValidBefore).Seconds()
	return float32(math.Max(0, left*100/total))
}

func (l *CertInfo) AbsoluteTimeUntilExpiry() time.Duration {
	return time.Until(l.ValidBefore)
}

type SshSignatureRequest struct {
	PublicKey  string
	Ttl        string
	Principals []string
	Extensions map[string]string

	VaultRole string
}

func ParseSshCertData(pubKeyBytes []byte) (CertInfo, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyBytes)
	if err != nil {
		return CertInfo{}, err
	}

	cert, ok := pubKey.(*ssh.Certificate)
	if !ok {
		return CertInfo{}, fmt.Errorf("pub key is not a valid certificate: %w", err)
	}

	return CertInfo{
		Type:        cert.Type(),
		Serial:      cert.Serial,
		ValidBefore: time.Unix(int64(cert.ValidBefore), 0).UTC(), //#nosec:G115
		ValidAfter:  time.Unix(int64(cert.ValidAfter), 0).UTC(),  //#nosec:G115
	}, nil
}

func convertUserKeyRequest(req SshSignatureRequest) map[string]any {
	data := map[string]interface{}{
		"public_key": req.PublicKey,
		"cert_type":  "user",
	}

	if len(req.Ttl) > 0 {
		data["ttl"] = req.Ttl
	}

	if len(req.Principals) > 0 {
		data["valid_principals"] = strings.Join(req.Principals, ",")
	}

	return data
}

func ReadSshCertFromDisk(publicKeyFile string) (CertInfo, error) {
	bytes, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return CertInfo{}, fmt.Errorf("reading cert failed: %w", err)
	}

	lifetime, err := ParseSshCertData(bytes)
	if err != nil {
		return CertInfo{}, fmt.Errorf("could not determine lifetime of cert: %w", err)
	}

	return lifetime, nil
}
