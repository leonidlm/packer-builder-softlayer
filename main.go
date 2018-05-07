package main

import (
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/watson-platform/packer-builder-softlayer/builder/softlayer"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(softlayer.Builder))
	server.Serve()
}
