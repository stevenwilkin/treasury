package main

import (
	"context"
	"net"
	"net/http"
)

const (
	socketPath = "/tmp/treasuryd.sock"
)

var client = http.Client{
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	},
}

func main() {
	rootCmd.Execute()
}
