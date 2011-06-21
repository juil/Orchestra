package main

import (
	"path"
	"os"
	"bufio"
	o "orchestra"
	"strings"
)

func pathFor(shortname string) (fullpath string) {
	return path.Clean(path.Join(*ConfigDirectory, shortname))
}

func ConfigLoad() {
	pfh, err := os.Open(pathFor("players"))
	o.MightFail("Couldn't open \"players\"", err)

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
