// config.go
//
// configuration file handling for orchestra.

package main

import (
	o "orchestra"
	"strings"
	"github.com/kuroneko/configureit"
	"crypto/tls"
	"os"
)

var configFile = configureit.New()

func init() {
	configFile.Add("x509 certificate", configureit.NewStringOption("/etc/orchestra/player_crt.pem"))
	configFile.Add("x509 private key", configureit.NewStringOption("/etc/orchestra/player_key.pem"))
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
}