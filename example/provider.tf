provider "mongodb" {
    url = "mongodb://localhost:27017"
    auth_database = "admin"
    auth_username = "db-admin-user"
    auth_password = var.admin_password
}