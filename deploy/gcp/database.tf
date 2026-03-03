# Cloud SQL PostgreSQL instance
resource "google_sql_database_instance" "staging" {
  name                = "ledgerspear-${var.environment}"
  database_version    = "POSTGRES_14"
  region              = var.region
  deletion_protection = false # Staging can be deleted

  depends_on = [google_service_networking_connection.private_vpc]

  settings {
    tier              = var.db_tier
    disk_size         = 10
    disk_autoresize   = false
    availability_type = "ZONAL" # Single zone for staging

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.vpc.id
    }

    backup_configuration {
      enabled = false # No backups for staging
    }
  }
}

# Database
resource "google_sql_database" "ledgerspear" {
  name     = "ledgerspear"
  instance = google_sql_database_instance.staging.name
}

# Database user
resource "google_sql_user" "ledgerspear" {
  name     = "ledgerspear"
  instance = google_sql_database_instance.staging.name
  password = var.db_password
}
