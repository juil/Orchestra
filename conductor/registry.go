/* registry.go
 *
*/

package main;

import (
	o "orchestra"
)

type registryRequest struct {
	operation		int
	hostname		string
	hostlist		[]string
	responseChannel		chan *registryResponse
}

type registryResponse struct {
	success			bool
	info			*ClientInfo
}

const (
	requestAdd = iota
	requestGet
	requestDelete
	requestSync
)

var (
	chanRegistryRequest	= make(chan *registryRequest)
 	clientList 		= make(map[string]*ClientInfo)
)

func regInternalAdd(hostname string) {
	o.Warn("Registry: New Host \"%s\"", hostname)
	clientList[hostname] = NewClientInfo()
	clientList[hostname].Player = hostname
}

func regInternalDel(hostname string) {
	o.Warn("Registry: Deleting Host \"%s\"", hostname)
	/* remove it from the registry */
	clientList[hostname] = nil, false
}

func manageRegistry() {
	for {
		req := <-chanRegistryRequest
		resp := new(registryResponse)
		/* by default, we failed. */
		resp.success = false
		switch (req.operation) {
		case requestAdd:
			_, exists := clientList[req.hostname]
			if exists {
				resp.success = false
			} else {
				regInternalAdd(req.hostname)
				resp.success = true
			}
		case requestGet:
			clinfo, exists := clientList[req.hostname]
			if exists {
				resp.success = true
				resp.info = clinfo
			} else {
				resp.success = false
			}
		case requestDelete:
			_, exists := clientList[req.hostname]
			if exists {
				resp.success = true
				regInternalDel(req.hostname)
			} else {
				resp.success = false
			}
		case requestSync:
			/* we need to make sure the registered clients matches
			 * the hostlist we're given.
			 *
			 * First, we transform the array into a map
			 */
			newhosts := make(map[string]bool)
			for k,_ := range req.hostlist {
				newhosts[req.hostlist[k]] = true
			}
			/* now, scan the current list, checking to see if
			 * they exist.  Remove them from the newhosts map
			 * if they do exist. 
			 */
			for k,_ := range clientList {
				_, exists := newhosts[k]
				if exists {
					/* remove it from the newhosts map */
					newhosts[k] = false, false
				} else {
					regInternalDel(k)
				}
			}
			/* now that we're finished, we should only have
			 * new clients in the newhosts list left. 
			 */
			for k,_ := range newhosts {
				regInternalAdd(k)
			}
			/* and we're done. */
		}
		if req.responseChannel != nil {
			req.responseChannel <- resp
		}
	}
}

func StartRegistry() {
	go manageRegistry()
}

func newRequest() (req *registryRequest) {
	req = new(registryRequest)
	req.responseChannel = make(chan *registryResponse, 1)

	return req
}
	
func ClientAdd(hostname string) (success bool) {
	r := newRequest()
	r.operation = requestAdd
	r.hostname = hostname
	chanRegistryRequest <- r
	resp := <- r.responseChannel
	
	return resp.success
}

func ClientDelete(hostname string) (success bool) {
	r := newRequest()
	r.operation = requestDelete
	r.hostname = hostname
	chanRegistryRequest <- r
	resp := <- r.responseChannel
	
	return resp.success
}

func ClientGet(hostname string) (info *ClientInfo) {
	r := newRequest()
	r.operation = requestGet
	r.hostname = hostname
	chanRegistryRequest <- r
	resp := <- r.responseChannel
	if resp.success {
		return resp.info
	}
	return nil
}

func ClientUpdateKnown(hostnames []string) {
	/* this is an asynchronous, we feed it into the registry 
	 * and it'll look after itself.
	*/
	r := newRequest()
	r.operation = requestSync
	r.hostlist = hostnames
	chanRegistryRequest <- r
}