package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func GenerateCACertificate() ([]byte, []byte, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{Organization: []string{"Dockport CA"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	certDer, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})

	caCert, err := x509.ParseCertificate(certDer)
	if err != nil {
		return nil, nil, err
	}

	os.WriteFile("data/tls/tls.cer", certPem, 0644)

	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	template = &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
	}

	certDer, _ = x509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, caKey)

	certPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})
	keyBytes, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		return nil, nil, err
	}

	keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	os.WriteFile("data/tls/tls.crt", certPem, 0644)
	os.WriteFile("data/tls/tls.key", keyPem, 0600)

	return certPem, keyPem, nil
}
