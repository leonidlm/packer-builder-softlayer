package main

import (
	"github.com/jamesdobson/packer-builder-softlayer/builder/softlayer"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(softlayer.Builder))
	server.Serve()
}
