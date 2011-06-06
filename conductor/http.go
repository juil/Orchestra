/* http.go
 *
 * HTTP status server.
*/

package main

import (
	"fmt"
	"http"
	"orchestra"
)

/* default ports are all in server.go */

func StartHTTP() {
	go httpServer()
}

func returnStatus(w http.ResponseWriter, r *http.Request) {
	tasks, players := DispatchStatus()
	fmt.Fprintf(w, "<p>Tasks Waiting: %d</p>\n", tasks)
	fmt.Fprintf(w, "<p>Players Idle:</p>\n<ul>\n")
	var i int
	for i = 0; i < len(players); i++ {
		fmt.Fprintf(w, "<li>%s</li>\n", players[i])
	}
	if (i == 0) {
		fmt.Fprintf(w, "<li>none</li>")
	}
	fmt.Fprintf(w, "</ul>")
}

func httpServer() {
	laddr := fmt.Sprintf(":%d", orchestra.DefaultHTTPPort)
	http.HandleFunc("/", returnStatus)
	http.ListenAndServe(laddr, nil)
}

