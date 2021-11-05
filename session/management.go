package session

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/pennz/amqp/utils"
)

// Session copied from https://pkg.go.dev/github.com/streadway/amqp#example-package
func myTLSConfig(chain string, fullchain string, privkey string) *tls.Config {
	var tlsConfig tls.Config
	tlsConfig.RootCAs = x509.NewCertPool()

	// If you use self generated CA root (chain), you can append it. Otherwise,
	// the default ones trusted can work? -> no, for real CA root, this step is still needed

	if ca, err := ioutil.ReadFile(chain); err == nil {
		if ok := tlsConfig.RootCAs.AppendCertsFromPEM(ca); ok == false {
			log.Fatalf("%s", "Failed to AppendCertsFromPEM")
		}
	} else {
		utils.FailOnError(err, "Failed to read PEM encoded certificates")
	}

	if cert, err := tls.LoadX509KeyPair(fullchain, privkey); err == nil {
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	} else {
		utils.FailOnError(err, "Failed to LoadX509KeyPair")
	}
	return &tlsConfig
}
