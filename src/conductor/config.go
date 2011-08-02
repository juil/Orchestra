package main

import (
	"os"
	"bufio"
	o "orchestra"
	"strings"
	c "github.com/kuroneko/configureit"
)

var configFile *c.Config = c.New()

func init() {
	configFile.Add("x509 certificate", c.NewStringOption("/etc/conductor/conductor_crt.pem"))
	configFile.Add("x509 private key", c.NewStringOption("/etc/conductor/conductor_key.pem"))
	configFile.Add("bind address", c.NewStringOption(""))
	configFile.Add("server name", c.NewStringOption(""))
	configFile.Add("audience socket path", c.NewStringOption("/var/run/conductor.sock"))
	configFile.Add("conductor state path", c.NewStringOption("/var/spool/orchestra"))
	configFile.Add("player file path", c.NewStringOption("/etc/conductor/players"))
}

func GetStringOpt(key string) string {
	cnode := configFile.Get(key)
	if cnode == nil {
		o.Assert("tried to get a configuration option that doesn't exist.")
	}
	sopt, ok := cnode.(*c.StringOption)
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
