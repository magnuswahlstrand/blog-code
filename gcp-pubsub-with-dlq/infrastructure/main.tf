terraform {
}

locals {
  topic_name  = "user-created"
  app_name    = "app"
}

provider "google" {
  project = var.project_id
}

module "pubsub-main" {
  source     = "terraform-google-modules/pubsub/google"
  project_id = var.project_id

  topic              = "${local.topic_name}"
  pull_subscriptions = [
    {
      name                  = "${local.app_name}.${local.topic_name}"
      dead_letter_topic     = module.pubsub-dlq.id
      max_delivery_attempts = 5
      maximum_backoff       = "10s"
      minimum_backoff       = "1s"
    }
  ]
}

module "pubsub-dlq" {
  source     = "terraform-google-modules/pubsub/google"
  project_id = var.project_id

  topic              = "${local.app_name}.${local.topic_name}.dlq"
  pull_subscriptions = [
    {
      name = "${local.app_name}.${local.topic_name}.dlq"
    }
  ]
}

resource "google_monitoring_alert_policy" "alert_policy" {
  display_name = "Messages on dead-letter queue (app: ${local.app_name}, topic: ${local.topic_name})"
  combiner     = "OR"
  conditions {
    display_name = "Cloud Pub/Sub Subscription - Unacked messages"
    condition_threshold {
      filter     = "resource.type = \"pubsub_subscription\" AND resource.labels.subscription_id = \"${module.pubsub-dlq.subscription_names.0}\" AND metric.type = \"pubsub.googleapis.com/subscription/num_undelivered_messages\""
      duration   = "120s"
      comparison = "COMPARISON_GT"
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }
  notification_channels = [
    "projects/b32-demo-projects/notificationChannels/6746093406737316903"
  ]

  user_labels = {
    app : local.app_name
  }
}
