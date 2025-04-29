provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "redis.googleapis.com",
    "cloudbuild.googleapis.com",
    "compute.googleapis.com",
    "servicenetworking.googleapis.com",
    "vpcaccess.googleapis.com",
    "secretmanager.googleapis.com",
    "cloudkms.googleapis.com",
    "iamcredentials.googleapis.com",
    "storage.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "dns.googleapis.com"
  ])
  
  service            = each.key
  disable_on_destroy = false
}

# Create service account for Cloud Run
resource "google_service_account" "cloudrun_sa" {
  account_id   = "redora-cloudrun"
  display_name = "Redora Cloud Run Service Account"
  depends_on   = [google_project_service.required_apis]
}