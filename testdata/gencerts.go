package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// generateKey generates a new RSA private key
func generateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// saveCertificate saves a certificate to a file in PEM format
func saveCertificate(path string, derBytes []byte) error {
	certOut, err := os.Create(path)
	if err != nil {
		return err
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}
	return nil
}

// savePrivateKey saves a private key to a file in PKCS8 format
func savePrivateKey(path string, key *rsa.PrivateKey) error {
	keyOut, err := os.Create(path)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes})
	if err != nil {
		return err
	}
	return nil
}

// generateSelfSignedCert generates a self-signed certificate
func generateSelfSignedCert(commonName string, isCA bool, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := generateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"SMIME Gateway Test"},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	// For TLS server, add localhost and 127.0.0.1 as SANs
	if commonName == "TLS Server" {
		template.DNSNames = []string{"localhost"}
		template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, key, nil
}

// generateSignedCert generates a certificate signed by a CA
func generateSignedCert(commonName string, email string, caCert *x509.Certificate, caKey *rsa.PrivateKey, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := generateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"SMIME Gateway Test"},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}

	if email != "" {
		template.EmailAddresses = []string{email}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, key, nil
}

func main() {
	// We're already in the testdata directory, so no need to create it

	// 1. Generate self-signed TLS server certificate
	fmt.Println("Generating TLS server certificate...")
	serverCert, serverKey, err := generateSelfSignedCert("TLS Server", false, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	if err != nil {
		fmt.Printf("Error generating TLS server certificate: %v\n", err)
		return
	}

	err = saveCertificate("tlsserver.cer", serverCert.Raw)
	if err != nil {
		fmt.Printf("Error saving TLS server certificate: %v\n", err)
		return
	}

	err = savePrivateKey("tlsserver.pkcs8", serverKey)
	if err != nil {
		fmt.Printf("Error saving TLS server private key: %v\n", err)
		return
	}

	// 2. Generate self-signed TLS client certificate
	fmt.Println("Generating TLS client certificate...")
	clientCert, clientKey, err := generateSelfSignedCert("TLS Client", false, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	if err != nil {
		fmt.Printf("Error generating TLS client certificate: %v\n", err)
		return
	}

	err = saveCertificate("tlsclient.cer", clientCert.Raw)
	if err != nil {
		fmt.Printf("Error saving TLS client certificate: %v\n", err)
		return
	}

	err = savePrivateKey("tlsclient.pkcs8", clientKey)
	if err != nil {
		fmt.Printf("Error saving TLS client private key: %v\n", err)
		return
	}

	// 3. Generate self-signed CA for user certificates
	fmt.Println("Generating user CA certificate...")
	caCert, caKey, err := generateSelfSignedCert("User CA", true, x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign, nil)
	if err != nil {
		fmt.Printf("Error generating user CA certificate: %v\n", err)
		return
	}

	err = saveCertificate("userca.cer", caCert.Raw)
	if err != nil {
		fmt.Printf("Error saving user CA certificate: %v\n", err)
		return
	}

	err = savePrivateKey("userca.pkcs8", caKey)
	if err != nil {
		fmt.Printf("Error saving user CA private key: %v\n", err)
		return
	}

	// 4. Generate S/MIME certificates for Alice and Bob
	// Alice's encryption certificate
	fmt.Println("Generating Alice's S/MIME certificates...")
	aliceEncCert, aliceEncKey, err := generateSignedCert("Alice", "alice@example.com", caCert, caKey, x509.KeyUsageKeyEncipherment, []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection})
	if err != nil {
		fmt.Printf("Error generating Alice's encryption certificate: %v\n", err)
		return
	}

	err = saveCertificate("alice_enc.cer", aliceEncCert.Raw)
	if err != nil {
		fmt.Printf("Error saving Alice's encryption certificate: %v\n", err)
		return
	}

	err = savePrivateKey("alice_enc.pkcs8", aliceEncKey)
	if err != nil {
		fmt.Printf("Error saving Alice's encryption private key: %v\n", err)
		return
	}

	// Alice's signing certificate
	aliceSignCert, aliceSignKey, err := generateSignedCert("Alice", "alice@example.com", caCert, caKey, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection})
	if err != nil {
		fmt.Printf("Error generating Alice's signing certificate: %v\n", err)
		return
	}

	err = saveCertificate("alice_sign.cer", aliceSignCert.Raw)
	if err != nil {
		fmt.Printf("Error saving Alice's signing certificate: %v\n", err)
		return
	}

	err = savePrivateKey("alice_sign.pkcs8", aliceSignKey)
	if err != nil {
		fmt.Printf("Error saving Alice's signing private key: %v\n", err)
		return
	}

	// Bob's encryption certificate
	fmt.Println("Generating Bob's S/MIME certificates...")
	bobEncCert, bobEncKey, err := generateSignedCert("Bob", "bob@example.com", caCert, caKey, x509.KeyUsageKeyEncipherment, []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection})
	if err != nil {
		fmt.Printf("Error generating Bob's encryption certificate: %v\n", err)
		return
	}

	err = saveCertificate("bob_enc.cer", bobEncCert.Raw)
	if err != nil {
		fmt.Printf("Error saving Bob's encryption certificate: %v\n", err)
		return
	}

	err = savePrivateKey("bob_enc.pkcs8", bobEncKey)
	if err != nil {
		fmt.Printf("Error saving Bob's encryption private key: %v\n", err)
		return
	}

	// Bob's signing certificate
	bobSignCert, bobSignKey, err := generateSignedCert("Bob", "bob@example.com", caCert, caKey, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection})
	if err != nil {
		fmt.Printf("Error generating Bob's signing certificate: %v\n", err)
		return
	}

	err = saveCertificate("bob_sign.cer", bobSignCert.Raw)
	if err != nil {
		fmt.Printf("Error saving Bob's signing certificate: %v\n", err)
		return
	}

	err = savePrivateKey("bob_sign.pkcs8", bobSignKey)
	if err != nil {
		fmt.Printf("Error saving Bob's signing private key: %v\n", err)
		return
	}

	fmt.Println("All certificates and keys generated successfully in the testdata directory.")
}
