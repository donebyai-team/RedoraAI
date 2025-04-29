variable "project_id" {
  description = "The GCP project ID"
  default     = "redora"
}

variable "region" {
  description = "The GCP region"
  default     = "us-east1"
}

variable "zone" {
  description = "The GCP zone"
  default     = "us-east1-b"
}

variable "github_repo" {
  description = "GitHub repository in format owner/repo"
  default     = "shank318/doota"
}

variable "domain" {
  description = "Base domain for the application"
  default     = "donebyai.team"
}

variable "developers_group" {
  description = "Google Group for developers"
  default     = "gcp-developers@donebyai.team"
}

variable "db_password_length" {
  description = "Length of the generated database password"
  default     = 32
}