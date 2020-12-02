package mongodb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
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
type RunCommandOutput struct {
	Ok int64 `bson:"ok" validate:"required"`
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
			"authentication_restrictions": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: false,
				Optional: true,
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
	authenticationRestrictions := d.Get("authentication_restrictions").(string)
	mongodbRoles := getMongoDBUserRoles(roles, dbname)
	mongodbAuthRestrictions := getMongoDBAuthRestrictions(authenticationRestrictions)

	var result bson.M
	// Connect to mongodb using environment variables
	cmd := bson.D{
		primitive.E{Key: "createUser", Value: username},
		primitive.E{Key: "pwd", Value: password},
		primitive.E{Key: "roles", Value: mongodbRoles},
		primitive.E{Key: "authenticationRestrictions", Value: mongodbAuthRestrictions},
	}
	err := client.Database(dbname).RunCommand(context.Background(), cmd).Decode(&result)

	// Unmarshal into Result struct
	var c *RunCommandOutput
	data, _ := bson.Marshal(result)
	bson.Unmarshal(data, &c)

	if c.Ok != 1 {
		return fmt.Errorf("Failed to create user: %s. Role: %v. Error: %s", username, mongodbRoles, err)
	}

	return readMongoDBUser(d, meta)
}

func resourceMongoDBUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mongo.Client)

	dbname := d.Get("database").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	roles := d.Get("roles").(*schema.Set)
	mongodbRoles := getMongoDBUserRoles(roles, dbname)
	authenticationRestrictions := d.Get("authentication_restrictions").(string)
	mongodbAuthRestrictions := getMongoDBAuthRestrictions(authenticationRestrictions)

	output := client.Database(dbname).RunCommand(context.Background(), bson.D{
		{"updateUser", username},
		{"pwd", password},
		{"roles", mongodbRoles},
		{"authenticationRestrictions", mongodbAuthRestrictions},
	})

	if output.Err() != nil && output.Err() != mongo.ErrNoDocuments {
		return fmt.Errorf("Failed to update user: %s. Error: %#v", username, output.Err())
	}

	return readMongoDBUser(d, meta)
}

func resourceMongoDBUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mongo.Client)

	dbname := d.Get("database").(string)
	username := d.Get("username").(string)

	var result bson.M
	err := client.Database(dbname).RunCommand(context.Background(), bson.D{
		{"dropUser", username},
	}).Decode(&result)

	// Unmarshal into Result struct
	var c *RunCommandOutput
	data, _ := bson.Marshal(result)
	bson.Unmarshal(data, &c)

	if c.Ok != 1 {
		return fmt.Errorf("Failed to drop user: %s. Error: %s", username, err)
	}

	return nil
}

func resourceMongoDBUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*mongo.Client)

	dbname := d.Get("database").(string)
	username := d.Get("username").(string)

	var result bson.M
	err := client.Database(dbname).RunCommand(context.Background(), bson.D{
		{"usersInfo", username},
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

func getMongoDBUserRoles(roles *schema.Set, database string) []bson.D {

	rolesDocs := make([]bson.D, 0)

	for _, role := range roles.List() {

		// Generate Mongo role type bson.D
		type r struct {
			Role string
			DB   string
		}
		data, _ := bson.Marshal(r{role.(string), database})

		// Unmarshal bson into bson.D document
		var doc bson.D
		bson.Unmarshal(data, &doc)
		rolesDocs = append(rolesDocs, doc)
	}

	return rolesDocs
}

func getMongoDBAuthRestrictions(authRestrictions string) []bson.D {
	restrictionDocs := make([]bson.D, 0)

	type r struct {
		ClientSource  []string `json:"clientSource"`
		ServerAddress []string `json:"serverAddress"`
	}

	var restrictions []r
	err := json.Unmarshal([]byte(authRestrictions), &restrictions)
	if err != nil {
		log.Fatalf("Unable to unmarshal `authenticationRestrictions` JSON: %e", err)
	}

	for _, restriction := range restrictions {

		r := map[string][]string{
			"clientSource":  restriction.ClientSource,
			"serverAddress": restriction.ServerAddress,
		}

		var doc bson.D

		data, err := bson.Marshal(r)
		if err != nil {
			log.Fatalf("Unable to marshal `authenticationRestrictions` to BSON: %e", err)
		}
		err = bson.Unmarshal(data, &doc)

		restrictionDocs = append(restrictionDocs, doc)
	}
	return restrictionDocs
}
