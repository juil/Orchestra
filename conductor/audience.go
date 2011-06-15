/* audience.go
*/

package main

import (
	"os"
	"io"
	"json"
	"net"
)

func handleAudienceRequest(c net.Conn) {
	defer c.Close()

	c.SetTimeout(0)
	r, _ := c.(io.Reader)
	w, _ := c.(io.Writer)
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)

	outobj = make(map[string]interface{})
	err := dec.Decode(outobj)
	if err != nil {
		o.Warn("Error decoding JSON talking to audience: %s", err)
		return
	}

	opObj, found := outobj["op"]
	if !found || op == nil {
		resp := &[]interface{} [ string("Malformed Request"), nil ]
		enc.Encode(resp)
		return
	}
	op, ok := opObj.(string)
	if !ok {
		resp := &[]interface{} [ string("Malformed Request"), nil ]
		enc.Encode(resp)
		return		
	}

	
}

func NewAudienceListener(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			o.Warn("Accept() failed on Audience Listenter.")
			break
		}
		go handleAudienceRequest(c)
	}
}

