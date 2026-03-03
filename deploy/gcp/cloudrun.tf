# Cloud Run service for the Go backend
resource "google_cloud_run_v2_service" "backend" {
  name     = "ledgerspear-api"
  location = var.region

  template {
    containers {
      image = var.backend_image != "" ? var.backend_image : "${var.region}-docker.pkg.dev/${var.project_id}/ledgerspear/backend:latest"

      ports {
        container_port = 8080
      }

      # Database configuration
      env {
        name  = "DB_HOST"
        value = google_sql_database_instance.staging.private_ip_address
      }
      env {
        name  = "DB_PORT"
        value = "5432"
      }
      env {
        name  = "DB_USER"
        value = "ledgerspear"
      }
      env {
        name  = "DB_NAME"
        value = "ledgerspear"
      }
      env {
        name  = "DB_SSLMODE"
        value = "require"
      }
      env {
        name  = "DB_MIGRATIONS_PATH"
        value = "/app/migrations"
      }

      # Server configuration
      env {
        name  = "SERVER_PORT"
        value = "8080"
      }

      # Secrets from Secret Manager
      env {
        name = "DB_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.db_password.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "ENCRYPTION_MASTER_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.encryption_key.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "SHOPIFY_CLIENT_ID"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.shopify_client_id.secret_id
            version = "latest"
          }
        }
      }
      env {
        name = "SHOPIFY_CLIENT_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.shopify_client_secret.secret_id
            version = "latest"
          }
        }
      }

      # Firebase credentials mounted as secret volume
      env {
        name  = "FIREBASE_CREDENTIALS_FILE"
        value = "/secrets/firebase/credentials.json"
      }

      volume_mounts {
        name       = "firebase-creds"
        mount_path = "/secrets/firebase"
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      startup_probe {
        http_get {
          path = "/health"
        }
        initial_delay_seconds = 5
        period_seconds        = 3
      }
    }

    volumes {
      name = "firebase-creds"
      secret {
        secret = google_secret_manager_secret.firebase_credentials.secret_id
        items {
          version = "latest"
          path    = "credentials.json"
        }
      }
    }

    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    scaling {
      min_instance_count = 0 # Scale to zero when idle
      max_instance_count = 2 # Staging cap
    }
  }
}

# Allow unauthenticated access (public API)
resource "google_cloud_run_v2_service_iam_member" "public" {
  name     = google_cloud_run_v2_service.backend.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}
