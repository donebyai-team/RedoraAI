# Cloud Run service account permissions
resource "google_project_iam_member" "cloudrun_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

resource "google_project_iam_member" "cloudrun_secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

resource "google_project_iam_member" "cloudrun_redis" {
  project = var.project_id
  role    = "roles/redis.viewer"
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

# Developer team permissions
resource "google_project_iam_member" "developers_cloudrun_admin" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "group:${var.developers_group}"
}

resource "google_project_iam_member" "developers_cloudrun_developer" {
  project = var.project_id
  role    = "roles/run.developer"
  member  = "group:${var.developers_group}"
}

