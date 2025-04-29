resource "google_cloud_run_service" "backend" {
  name     = "redora-backend"
  location = var.region

  template {
    metadata {
      annotations = {
        "run.googleapis.com/vpc-access-connector" = google_vpc_access_connector.connector.id
      }
    }

    spec {
      service_account_name = google_service_account.cloudrun_sa.email

      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello" # Dummy image
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [
    google_project_service.required_apis,
    google_sql_database_instance.postgres,
    google_redis_instance.memorystore,
    google_vpc_access_connector.connector
  ]
}

resource "google_cloud_run_service" "frontend" {
  name     = "redora-frontend"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/cloudrun/hello" # Dummy image
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_vpc_access_connector" "connector" {
  name          = "redora-vpc-connector"
  region        = var.region
  network       = google_compute_network.redora_vpc.name
  ip_cidr_range = "10.20.0.0/28"
  machine_type  = "e2-micro"
  min_instances = 2
  max_instances = 3
}



