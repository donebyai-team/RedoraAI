terraform {
  backend "gcs" {
    bucket = "redora-terraform-state"
    prefix = "terraform/state"
  }
}