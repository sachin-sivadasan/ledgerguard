# Secret Manager entries for sensitive configuration

resource "google_secret_manager_secret" "db_password" {
  secret_id = "ledgerspear-db-password"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = var.db_password
}

resource "google_secret_manager_secret" "firebase_credentials" {
  secret_id = "ledgerspear-firebase-credentials"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "encryption_key" {
  secret_id = "ledgerspear-encryption-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "shopify_client_id" {
  secret_id = "ledgerspear-shopify-client-id"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "shopify_client_secret" {
  secret_id = "ledgerspear-shopify-client-secret"
  replication {
    auto {}
  }
}

# IAM: Allow Cloud Run service account to access secrets
data "google_project" "project" {}

resource "google_secret_manager_secret_iam_member" "db_password_access" {
  secret_id = google_secret_manager_secret.db_password.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "firebase_access" {
  secret_id = google_secret_manager_secret.firebase_credentials.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "encryption_access" {
  secret_id = google_secret_manager_secret.encryption_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "shopify_id_access" {
  secret_id = google_secret_manager_secret.shopify_client_id.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "shopify_secret_access" {
  secret_id = google_secret_manager_secret.shopify_client_secret.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}
