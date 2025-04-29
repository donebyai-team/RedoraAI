resource "google_iam_workload_identity_pool" "default" {
  workload_identity_pool_id = "default"
}


resource "google_iam_workload_identity_pool_provider" "github-actions" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.default.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-actions"

  attribute_mapping = {
    "google.subject"             = "assertion.sub"            # GitHub subject (user/org)
    "attribute.actor"            = "assertion.actor"          # GitHub actor (user who triggered the action)
    "attribute.repository"       = "assertion.repository"     # GitHub repository name
    "attribute.repository_owner" = "assertion.repository_owner"  # GitHub repository owner
  }

  // The missing attribute condition in common expression langauge:
  attribute_condition = "attribute.actor == assertion.actor && attribute.repository == assertion.repository && attribute.repository_owner == assertion.repository_owner"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account" "github_actions" {
  account_id   = "redora-github-actions"
  display_name = "Redora GitHub Actions Service Account"
}

resource "google_project_iam_member" "github_actions_cloudbuild" {
  project = var.project_id
  role    = "roles/cloudbuild.builds.editor"
  member  = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_project_iam_member" "github_actions_cloudrun" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_project_iam_member" "github_actions_storage" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_project_iam_member" "github_actions_service_account_user" {
  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_service_account_iam_binding" "github_actions_iam_binding" {
  service_account_id = google_service_account.github_actions.name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.default.name}/attribute.repository_owner/shank318",  # GitHub username for the organization
  ]
}
