resource "google_cloud_run_domain_mapping" "frontend" {
  location = var.region
  name     = "app.${var.domain}"

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_service.frontend.name
  }

  depends_on = [google_cloud_run_service.frontend]
}

resource "google_cloud_run_domain_mapping" "backend" {
  location = var.region
  name     = "api.${var.domain}"

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_service.backend.name
  }

  depends_on = [google_cloud_run_service.backend]
}