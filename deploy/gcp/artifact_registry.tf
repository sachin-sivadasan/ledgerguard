# Docker image repository
resource "google_artifact_registry_repository" "backend" {
  location      = var.region
  repository_id = "ledgerspear"
  format        = "DOCKER"
  description   = "LedgerSpear backend Docker images"
}
