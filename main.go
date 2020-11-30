package main

import (
	"github.com/everonhq/terraform-provider-mongodb/mongodb"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mongodb.Provider})
}
