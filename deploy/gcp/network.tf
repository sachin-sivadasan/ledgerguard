# VPC for private networking between Cloud Run and Cloud SQL
resource "google_compute_network" "vpc" {
  name                    = "ledgerspear-${var.environment}-vpc"
  auto_create_subnetworks = true
}

# Reserve IP range for Cloud SQL private services
resource "google_compute_global_address" "private_ip_range" {
  name          = "ledgerspear-${var.environment}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
}

# Private services connection for Cloud SQL
resource "google_service_networking_connection" "private_vpc" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_range.name]
}

# VPC Access Connector for Cloud Run → Cloud SQL
resource "google_vpc_access_connector" "connector" {
  name          = "cloudrun-sql-connector"
  region        = var.region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.8.0.0/28"
  machine_type  = "e2-micro"
  min_instances = 2
  max_instances = 3
}
