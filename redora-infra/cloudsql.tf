resource "google_sql_database_instance" "postgres" {
  name             = "redora-postgres"
  database_version = "POSTGRES_14"
  region           = var.region

  settings {
    tier              = "db-f1-micro"
    availability_type = "ZONAL"
    disk_size         = 10
    disk_type         = "PD_SSD"

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.redora_vpc.id
    }

    backup_configuration {
      enabled = true
    }
  }

  deletion_protection = false
  depends_on = [
    google_service_networking_connection.private_vpc_connection,
    google_project_service.required_apis
  ]
}

resource "google_sql_database" "database" {
  name     = "dev-node"
  instance = google_sql_database_instance.postgres.name
}

resource "google_sql_user" "user" {
  name     = "dev-node"
  instance = google_sql_database_instance.postgres.name
  password = random_password.db_password.result
}

resource "google_compute_network" "redora_vpc" {
  name                    = "redora-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_global_address" "private_ip_address" {
  name          = "redora-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.redora_vpc.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.redora_vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}