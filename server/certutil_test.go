package server_test

/*
MIT License

Copyright (c) 2026 Shane

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"sync"
	"testing"
	"time"
)

// tlsCerts is the (cert, key) PEM pair minted once per test process. It is the
// Go stand-in for the test __utils__/tls tlsTestCerts helper, which generates a
// fresh self-signed cert per process instead of committing static fixtures.
type tlsCerts struct {
	cert []byte
	key  []byte
}

var (
	tlsCertOnce  sync.Once
	tlsCertValue tlsCerts
	tlsCertErr   error
)

// tlsTestCerts mints (and memoizes) a self-signed RSA certificate for
// commonName "localhost", returning the PEM-encoded cert and key.
func tlsTestCerts(t *testing.T) tlsCerts {
	t.Helper()
	tlsCertOnce.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			tlsCertErr = err
			return
		}
		tmpl := &x509.Certificate{
			SerialNumber:          big.NewInt(1),
			Subject:               pkix.Name{CommonName: "localhost"},
			NotBefore:             time.Now().Add(-time.Hour),
			NotAfter:              time.Now().Add(24 * time.Hour),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			DNSNames:              []string{"localhost"},
			IsCA:                  true,
			BasicConstraintsValid: true,
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
		if err != nil {
			tlsCertErr = err
			return
		}
		tlsCertValue = tlsCerts{
			cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
			key:  pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
		}
	})
	if tlsCertErr != nil {
		t.Fatalf("mint self-signed cert: %v", tlsCertErr)
	}
	return tlsCertValue
}
