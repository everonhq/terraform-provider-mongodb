provider "mongodb" {
    url = "mongodb://localhost:27017"
}

resource "mongodb_user" "user" {
    database = "test"
    username = "user"
    password = "pass"
    roles = ["read", "dbAdmin", "userAdmin"]
}