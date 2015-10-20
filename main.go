package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/picocandy/packer-builder-softlayer/builder/softlayer"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(softlayer.Builder))
	server.Serve()
}
