Terraform Provider
==================

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/everonhq/terraform-provider-mongodb`

```sh
$ mkdir -p $GOPATH/src/github.com/everonhq; cd $GOPATH/src/github.com/everonhq
$ git clone git@github.com:everonhq/terraform-provider-mongodb
```

Enter the provider directory, create the go.mod and build the provider

```sh
$ cd $GOPATH/src/github.com/everonhq/terraform-provider-mongodb
$ go mod init
$ make build
```

Install provider into Terraform plugins dir
```sh
$ cp ~/go/bin/terraform-provider-mongodb ~/.terraform.d/plugins/darwin_amd64/
```

Once code updated, install provider, and run example tf:
```sh
cd ..
make build
cp ~/go/bin/terraform-provider-mongodb ~/.terraform.d/plugins/darwin_amd64/
cd example
terraform init
terraform apply
```

Using the provider
----------------------

### Provider with auth enabled:
```
provider "mongodb" {
    url = "mongodb://localhost:27017"
    auth_database = "admin"
    auth_username = "db-admin-user"
    auth_password = var.admin_password
}
```

Then, create `terraform.tfvars` with the password variable defined:
```
admin_password = "<provider login passsword>"
```

### Provider without auth:
```
provider "mongodb" {
    url = "mongodb://localhost:27017/test"
}

resource "mongodb_user" "user" {
    database = "test"
    username = "user"
    password = "pass"
    roles = ["read", "dbAdmin", "userAdmin"]
    authentication_restrictions = jsonencode([{
                                                 clientSource  = ["127.0.0.1"]
                                                 serverAddress = ["127.0.0.1"]
                                              }])
}
```

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-mongodb
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
