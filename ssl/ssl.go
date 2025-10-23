package ssl

import (
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

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

// GenerateSelfSignedCert генерирует самоподписанный сертификат
func GenerateSelfSignedCert(host string) (*tls.Certificate, error) {
	// Создаем приватный ключ
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Создаем сертификат
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"LCG Server"},
			Country:       []string{"RU"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost", host},
	}

	// Подписываем сертификат
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	// Создаем TLS сертификат
	cert := &tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}

	return cert, nil
}

// SaveCertToFile сохраняет сертификат и ключ в файлы
func SaveCertToFile(cert *tls.Certificate, certFile, keyFile string) error {
	// Создаем директорию если не существует
	certDir := filepath.Dir(certFile)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create cert directory: %v", err)
	}

	// Сохраняем сертификат
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to open cert file: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}); err != nil {
		return fmt.Errorf("failed to encode cert: %v", err)
	}

	// Сохраняем приватный ключ
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to open key file: %v", err)
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		return fmt.Errorf("failed to encode private key: %v", err)
	}

	return nil
}

// LoadOrGenerateCert загружает существующий сертификат или генерирует новый
func LoadOrGenerateCert(host string) (*tls.Certificate, error) {
	// Определяем пути к файлам сертификата
	certFile := config.AppConfig.Server.SSLCertFile
	keyFile := config.AppConfig.Server.SSLKeyFile

	// Если пути не указаны, используем стандартные
	if certFile == "" {
		certFile = filepath.Join(config.AppConfig.Server.ConfigFolder, "server", "ssl", "cert.pem")
	}
	if keyFile == "" {
		keyFile = filepath.Join(config.AppConfig.Server.ConfigFolder, "server", "ssl", "key.pem")
	}

	// Проверяем существующие файлы
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			// Загружаем существующий сертификат
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err == nil {
				return &cert, nil
			}
		}
	}

	// Генерируем новый сертификат
	cert, err := GenerateSelfSignedCert(host)
	if err != nil {
		return nil, err
	}

	// Сохраняем сертификат
	if err := SaveCertToFile(cert, certFile, keyFile); err != nil {
		return nil, err
	}

	return cert, nil
}

// IsSecureHost проверяет, является ли хост безопасным для HTTP
func IsSecureHost(host string) bool {
	secureHosts := []string{"localhost", "127.0.0.1", "::1"}
	for _, secureHost := range secureHosts {
		if host == secureHost {
			return true
		}
	}
	return false
}

// ShouldUseHTTPS определяет, нужно ли использовать HTTPS
func ShouldUseHTTPS(host string) bool {
	// Если хост не localhost/127.0.0.1, принуждаем к HTTPS
	if !IsSecureHost(host) {
		return true
	}

	// Если явно разрешен HTTP, используем HTTP
	if config.AppConfig.Server.AllowHTTP {
		return false
	}

	// По умолчанию для localhost используем HTTP
	return false
}
