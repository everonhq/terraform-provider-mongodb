package mongodb

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MONGODB_URL", ""),
				Description: "The MongoDB url.",
			},
			"auth_database": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MONGODB_AUTH_DATABASE", ""),
				Description: "The MongoDB authentication database.",
			},
			"auth_username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MONGODB_AUTH_USERNAME", ""),
				Description: "The MongoDB login username.",
			},
			"auth_password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MONGODB_AUTH_PASSWORD", ""),
				Description: "The MongoDB login password.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"mongodb_user": resourceMongoDBUser(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		URL:          d.Get("url").(string),
		AuthDatabase: d.Get("auth_database").(string),
		AuthUsername: d.Get("auth_username").(string),
		AuthPassword: d.Get("auth_password").(string),
	}

	client, err := config.loadAndValidate()

	if err != nil {
		return nil, err
	}

	return client, nil
}
