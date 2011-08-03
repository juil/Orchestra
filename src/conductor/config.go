package main

import (
	"os"
	"bufio"
	o "orchestra"
	"strings"
	"github.com/kuroneko/configureit"
)

var configFile *configureit.Config = configureit.New()

func init() {
	configFile.Add("x509 certificate", configureit.NewStringOption("/etc/orchestra/conductor_crt.pem"))
	configFile.Add("x509 private key", configureit.NewStringOption("/etc/orchestra/conductor_key.pem"))
	configFile.Add("ca certificates", configureit.NewPathListOption(nil))
	configFile.Add("bind address", configureit.NewStringOption(""))
	configFile.Add("server name", configureit.NewStringOption(""))
	configFile.Add("audience socket path", configureit.NewStringOption("/var/run/conductor.sock"))
	configFile.Add("conductor state path", configureit.NewStringOption("/var/spool/orchestra"))
	configFile.Add("player file path", configureit.NewStringOption("/etc/orchestra/players"))
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

	playerpath := strings.TrimSpace(GetStringOpt("player file path"))
	pfh, err := os.Open(playerpath)
	o.MightFail(err, "Couldn't open \"%s\"", playerpath)

	pbr := bufio.NewReader(pfh)

	ahmap := make(map[string]bool)
	for err = nil; err == nil; {
		var lb		[]byte
		var prefix	bool

		lb, prefix, err = pbr.ReadLine()

		if nil == lb {
			break;
		}
		if prefix {
			o.Fail("ConfigLoad: Short Read (prefix only)!")
		}
		
		line := strings.TrimSpace(string(lb))
		if line == "" {
			continue;
		}
		if line[0] == '#' {
			continue;
		}
		ahmap[line] = true
	}
	// convert newAuthorisedHosts to a slice
	authorisedHosts := make([]string, len(ahmap))
	idx := 0
	for k,_ := range ahmap {
		authorisedHosts[idx] = k
		idx++
	}
	ClientUpdateKnown(authorisedHosts)
}


func HostAuthorised(hostname string) (r bool) {
	/* if we haven't loaded the configuration, nobody is authorised */
	ci := ClientGet(hostname)
	return ci != nil
}
