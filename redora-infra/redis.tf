resource "google_redis_instance" "memorystore" {
  name           = "redora-memorystore"
  tier           = "BASIC"
  memory_size_gb = 1
  region         = var.region

  authorized_network = google_compute_network.redora_vpc.id
  display_name  = "Redora Memorystore"

  lifecycle {
    prevent_destroy = false
  }

  depends_on = [
    google_project_service.required_apis,
    google_compute_network.redora_vpc
  ]
}
