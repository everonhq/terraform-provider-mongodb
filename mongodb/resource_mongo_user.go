package mongodb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/globalsign/mgo/bson"
	"github.com/hashicorp/terraform/helper/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo usersInfo output Result struct
type UsersInfo struct {
	Ok             int64 `bson:"ok" validate:"required"`
	UsersInfoUsers `bson:",inline"`
}
type UsersInfoUsers struct {
	Users []UsersInfoUserConfig `bson:"users" validate:"required"`
}
type UsersInfoUserConfig struct {
	User string `bson:"user" validate:"required"`
}

func resourceMongoDBUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceMongoDBUserCreate,
		Update: resourceMongoDBUserUpdate,
		Read:   resourceMongoDBUserRead,
		Exists: resourceMongoDBUserExists,
		Delete: resourceMongoDBUserDelete,
		Schema: map[string]*schema.Schema{
			"database": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"roles": &schema.Schema{
				Type:     schema.TypeSet,
				Required: false,
				ForceNew: false,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func readMongoDBUser(d *schema.ResourceData, meta interface{}) error {
	dbname := d.Get("database").(string)
	username := d.Get("username").(string)

	var id bytes.Buffer
	id.WriteString(dbname)
	id.WriteString(".")
	id.WriteString(username)

	d.SetId(id.String())

	return nil
}

func resourceMongoDBUserRead(d *schema.ResourceData, meta interface{}) error {
	database := d.Get("database").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	var id bytes.Buffer
	id.WriteString(database)
	id.WriteString(".")
	id.WriteString(username)

	d.SetId(id.String())
	d.Set("username", username)
	d.Set("password", password)

	return nil
}

func resourceMongoDBUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mongo.Client)

	dbname := d.Get("database").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	roles := d.Get("roles").(*schema.Set)
	mongodbRoles := getMongoDBUserRoles(roles)

	err := client.Database(dbname).RunCommand(context.Background(), bson.D{
		{"createUser", username},
		{"pwd", password},
		{"roles", mongodbRoles},
	})
	if err != nil {
		return fmt.Errorf("Failed to create user: %s", username)
	}

	return readMongoDBUser(d, meta)
}

func resourceMongoDBUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mongo.Client)

	dbname := d.Get("database").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	roles := d.Get("roles").(*schema.Set)
	mongodbRoles := getMongoDBUserRoles(roles)

	err := client.Database(dbname).RunCommand(context.Background(), bson.D{
		{"updateUser", username},
		{"pwd", password},
		{"roles", mongodbRoles},
	})
	if err != nil {
		return fmt.Errorf("Failed to update user: %s", username)
	}

	return readMongoDBUser(d, meta)
}

func resourceMongoDBUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mongo.Client)

	dbname := d.Get("database").(string)
	username := d.Get("username").(string)

	err := client.Database(dbname).RunCommand(context.Background(), bson.D{
		{"dropUser", username},
	})
	if err != nil {
		return fmt.Errorf("Failed to drop user: %s", username)
	}

	return nil
}

func resourceMongoDBUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*mongo.Client)

	dbname := d.Get("database").(string)
	username := d.Get("username").(string)
	roles := d.Get("roles").(*schema.Set)
	mongodbRoles := getMongoDBUserRoles(roles)

	var result bson.M
	err := client.Database(dbname).RunCommand(context.Background(), bson.D{
		{"usersInfo", username},
		{"roles", mongodbRoles},
	}).Decode(&result)

	// Unmarshal into Result struct
	var c *UsersInfo
	data, _ := bson.Marshal(result)
	bson.Unmarshal(data, &c)

	if len(c.Users) < 1 || c.Users[0].User != username {
		return false, fmt.Errorf("Username: %s was not found in list of users returned by MongoDB. Must create new user", username)
	}

	return err == nil, nil
}

func getMongoDBUserRoles(roles *schema.Set) []bson.D {

	rolesDocs := make([]bson.D, 0)

	for _, role := range roles.List() {

		// Marshal RoleConfig struct to bson
		data, err := bson.Marshal(role.(string))
		if err != nil {
			return nil
		}
		// Unmarshal bson into bson.D document
		var doc bson.D
		err = bson.Unmarshal(data, &doc)
		rolesDocs = append(rolesDocs, doc)
	}

	return rolesDocs
}
