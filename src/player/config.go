// config.go
//
// configuration file handling for orchestra.

package main

import (
	o "orchestra"
	"strings"
	"github.com/kuroneko/configureit"
	"crypto/tls"
	"crypto/x509"
	"os"
)

var configFile = configureit.New()

func init() {
	configFile.Add("x509 certificate", configureit.NewStringOption("/etc/orchestra/player_crt.pem"))
	configFile.Add("x509 private key", configureit.NewStringOption("/etc/orchestra/player_key.pem"))
	configFile.Add("ca certificates", configureit.NewPathListOption(nil))
	configFile.Add("master", configureit.NewStringOption("conductor"))
	configFile.Add("score directory", configureit.NewStringOption("/usr/lib/orchestra/scores"))
	configFile.Add("player name", configureit.NewStringOption(""))
}

func GetStringOpt(key string) string {
	cnode := configFile.Get(key)
	if cnode == nil {
		o.Assert("tried to get a configuration option that doesn't exist.")
	}
	sopt, ok := cnode.(*configureit.StringOption)
	if !ok {
		o.Assert("tried to get a non-string configuration option with GetStringOpt")
	}
	return strings.TrimSpace(sopt.Value)
}

func GetCACertList() []string {
	cnode := configFile.Get("ca certificiates")
	if cnode == nil {
		o.Assert("tried to get a configuration option that doesn't exist.")
	}
	plopt, _ := cnode.(*configureit.PathListOption)
	return plopt.Values
}

func ConfigLoad() {
	// attempt to open the configuration file.
	fh, err := os.Open(*ConfigFile)
	if nil == err {
		defer fh.Close()
		// reset the config File data, then reload it.
		configFile.Reset()
		ierr := configFile.Read(fh, 1)
		o.MightFail(ierr, "Couldn't parse configuration")
	} else {
		o.Warn("Couldn't open configuration file: %s.  Proceeding anyway.", err)
	}

	// load the x509 certificates
	x509CertFilename := GetStringOpt("x509 certificate")
	x509PrivateKeyFilename := GetStringOpt("x509 private key")
	CertPair, err = tls.LoadX509KeyPair(x509CertFilename, x509PrivateKeyFilename)
	o.MightFail(err, "Couldn't load certificates")

	// load the CA Certs
	CACertPool = x509.NewCertPool()
	caCertNames := GetCACertList()
	if caCertNames != nil {
		for _, filename := range caCertNames {
			fh, err := os.Open(filename)
			if err != nil {
				o.Warn("Whilst parsing CA certs, couldn't open %s: %s", filename, err)
				continue
			}
			defer fh.Close()
			fi, err := fh.Stat()
			o.MightFail(err, "Couldn't stat CA certificate file: %s", filename)
			data := make([]byte, fi.Size)
			fh.Read(data)
			CACertPool.AppendCertsFromPEM(data)
		}
	}
}