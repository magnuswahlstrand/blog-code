terraform {
}

locals {
  pubsub_name = "user-created"
}

provider "google" {
  project = var.project_id
}

#data "google_project" "project" {
#}

# Service accounts and IAM
resource "google_service_account" "pubsub_sa" {
  account_id = "pubsub-sa"
}

module "pubsub-main" {
  source     = "terraform-google-modules/pubsub/google"
  project_id = var.project_id

  topic              = "${local.pubsub_name}"
  pull_subscriptions = [
    {
      name                  = "${local.pubsub_name}-sub"
      dead_letter_topic     = module.pubsub-dlq.id
      service_account       = google_service_account.pubsub_sa.email
      max_delivery_attempts = 5
      maximum_backoff       = "10s"
      minimum_backoff       = "1s"
    }
  ]
}


module "pubsub-dlq" {
  source     = "terraform-google-modules/pubsub/google"
  project_id = var.project_id

  topic              = "${local.pubsub_name}-dlq"
  pull_subscriptions = [
    {
      name = "${local.pubsub_name}-dlq-sub"
    }
  ]
}
