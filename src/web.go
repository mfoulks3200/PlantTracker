package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var templates *template.Template

type Session struct {
	User                User
	Token               string
	PendingRedirectPath string
}

func RunWebserver() {

	LoadPages()

	http.HandleFunc("/", HandleRequest)

	//Start Standard HTTPS Server
	if GetConfigurationPath("webserver.port") == "443" {
		if !(FileExists("server.crt") && FileExists("server.key")) {
			CreateSelfSignCerts()
		}
		LogMessage("Starting webserver on port " + GetConfigurationPath("webserver.port"))
		log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", nil))
		go http.ListenAndServe(":80", http.HandlerFunc(RedirectClientToHTTPS))

		//Start HTTP Server
	} else if GetConfigurationPath("webserver.port") == "80" {
		LogMessage("Warning: HTTPS is required for multiple features, including the mobile scanner to function")
		LogMessage("Starting webserver on port " + GetConfigurationPath("webserver.port"))
		log.Fatal(http.ListenAndServe(":80", nil))

		//Start HTTPS Server on nonstandard Port
	} else {
		if !(FileExists("server.crt") && FileExists("server.key")) {
			CreateSelfSignCerts()
		}
		LogMessage("Starting webserver on port " + GetConfigurationPath("webserver.port"))
		log.Fatal(http.ListenAndServeTLS(":"+GetConfigurationPath("webserver.port"), "server.crt", "server.key", nil))
		go http.ListenAndServe(":80", http.HandlerFunc(RedirectClientToHTTPS))
	}
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	//id := r.URL.Path[len("/admin/passwordChange/"):]
}

//http.Redirect(w, r, "./login", 301)

func RedirectClientToHTTPS(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, target,
		// see comments below and consider the codes 308, 302, or 301
		http.StatusTemporaryRedirect)
}

func CreateSelfSignCerts() {
	LogMessage("Did not detect server.crt and server.key, creating self-sign certs now")

	//Generate System CA
	//THIS WILL CAUSE BROWSER WARNINGS
	//Supply your own certs from a recognized CA to prevent this
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization:  []string{"Plant Tracker Instance Self Signed Cert"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"North Carolina"},
			StreetAddress: []string{"With <3"},
			PostalCode:    []string{"27513"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	//Create private key for CA
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		LogFatal(fmt.Sprint(err))
	}

	//Create public key for CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		LogFatal(fmt.Sprint(err))
	}

	//PEM Encode public key
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	//PEM Encode private key
	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	//Actually start creating the SLL certs
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization:  []string{"Plant Tracker Instance Self Signed Cert"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"North Carolina"},
			StreetAddress: []string{"With <3"},
			PostalCode:    []string{"27513"},
		},
		IPAddresses:  []net.IP{net.ParseIP(GetConfigurationPath("webserver.ipAddr")), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	//Create server pub and priv keys
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		LogFatal(fmt.Sprint(err))
	}

	//Sign those keys with our CA's authority
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		LogFatal(fmt.Sprint(err))
	}

	//PEM Encode the new certs
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certFile, err := os.Create("server.crt")
	_, err = certFile.Write(certPEM.Bytes())
	certFile.Sync()

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	keyFile, err := os.Create("server.key")
	_, err = certFile.Write(certPrivKeyPEM.Bytes())
	keyFile.Sync()

	certFileTemp := strings.Split(ReadTextFile("server.crt"), "-----END CERTIFICATE-----")

	certFile, err = os.Create("server.key")
	_, err = certFile.Write([]byte(certFileTemp[1]))
	certFile.Sync()

}
