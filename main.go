package main

import (
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"

	vf "github.com/david-gurley/nomad-vf-plugin/device"
)

func main() {
	// Serve the plugin
	plugins.Serve(factory)
}

// factory returns a new instance of our example device plugin
func factory(log log.Logger) interface{} {
	return vf.NewPlugin(log)
}
