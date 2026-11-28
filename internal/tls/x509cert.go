package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Cert struct {
	CertFile string
	KeyFile  string
}

// certPaths возвращает пути к сертификату и ключу.
func certPaths() (certFile, keyFile string, err error) {
	if c := os.Getenv("TLS_CERT_FILE"); c != "" {
		k := os.Getenv("TLS_KEY_FILE")
		if k == "" {
			return "", "", fmt.Errorf("TLS_CERT_FILE set but TLS_KEY_FILE empty")
		}
		return c, k, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}
	base := filepath.Join(dir, "gophkeeper")
	return filepath.Join(base, "cert.pem"), filepath.Join(base, "private.pem"), nil
}

// CreateX509Cert создаёт самоподписанный сертификат, если его нет,
// и возвращает пути к файлам.
func CreateX509Cert() (*Cert, error) {
	certFile, keyFile, err := certPaths()
	if err != nil {
		return nil, err
	}

	exist1 := fileExists(certFile)
	exist2 := fileExists(keyFile)
	if exist1 != exist2 {
		return nil, fmt.Errorf("inconsistent cert state: cert=%v key=%v", exist1, exist2)
	}

	if !exist1 && !exist2 {
		if err := generate(certFile, keyFile); err != nil {
			return nil, err
		}
	}

	// проверяем, что пара валидна
	if _, err := tls.LoadX509KeyPair(certFile, keyFile); err != nil {
		return nil, fmt.Errorf("invalid cert/key pair: %w", err)
	}

	return &Cert{CertFile: certFile, KeyFile: keyFile}, nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func generate(certFile, keyFile string) error {
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate serial: %w", err)
	}

	cert := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Gophkeeper"},
			Country:      []string{"RU"},
		},
		DNSNames:              []string{"localhost", "server"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	var certPEM bytes.Buffer
	if err := pem.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return err
	}

	var keyPEM bytes.Buffer
	if err := pem.Encode(&keyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(certFile), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(certFile, certPEM.Bytes(), 0o644); err != nil {
		return err
	}
	return os.WriteFile(keyFile, keyPEM.Bytes(), 0o600)
}
