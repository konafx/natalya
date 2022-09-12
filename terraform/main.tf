# main.tf

provider "google" {
  project = var.project
  region  = var.region
  zone    = var.zone
}
