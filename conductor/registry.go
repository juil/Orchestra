/* registry.go
 *
*/

package main;

import (
	"os"
)

type registryRequest struct {
	operation		int
	hostname		string
	info			*ClientInfo
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
	requestReplace
)

var (
	ErrClientAlreadyConnected = os.NewError("Client is already connected")

	chanRegistryRequest = make(chan *registryRequest)
)

func manageRegistry() {
 	clientList := make(map[string]*ClientInfo)
	
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
				clientList[req.hostname] = req.info
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
				clientList[req.hostname] = nil, false
			} else {
				resp.success = false
			}
		case requestReplace:
			_, exists := clientList[req.hostname]
			if exists {
				resp.success = true
				clientList[req.hostname] = req.info
			} else {
				resp.success = false
			}
		}
		req.responseChannel <- resp
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
	
func ClientAdd(hostname string, info *ClientInfo) (success bool) {
	r := newRequest()
	r.operation = requestAdd
	r.hostname = hostname
	r.info = info
	chanRegistryRequest <- r
	resp := <- r.responseChannel
	
	return resp.success
}

func ClientReplace(hostname string, info *ClientInfo) (success bool) {
	r := newRequest()
	r.operation = requestReplace
	r.hostname = hostname
	r.info = info
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