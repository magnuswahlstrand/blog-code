terraform {
}

locals {
}

provider "google" {
  project = var.project_id
}

locals {
  root_dir = abspath("../")
}

variable "gcp_service_list" {
  description = "The list of apis necessary for the project"
  type        = list(string)
  default     = [
    "cloudfunctions.googleapis.com",
    "cloudbuild.googleapis.com",
    "cloudscheduler.googleapis.com",
    "addressvalidation.googleapis.com"
  ]
}

resource "google_project_service" "gcp_services" {
  for_each = toset(var.gcp_service_list)
  project  = var.project_id
  service  = each.key
}

module "scheduled-function" {
  source                    = "terraform-google-modules/scheduled-function/google"
  version                   = "2.5.1"
  project_id                = var.project_id
  job_name                  = "scheduled_function"
  job_schedule              = "0 0 * * *"
  function_name             = "magnus_scheduled_function2"
  region                    = "europe-west1"
  function_source_directory = local.root_dir
  function_runtime          = "go120"
  function_entry_point      = "HelloPubSub"
  depends_on                = [
    google_project_service.gcp_services
  ]
}

resource "google_storage_bucket" "my_upload_bucket" {
  name     = "magnus_upload_bucket"
  location = "EU"
}

module "trigger-function" {
  source        = "terraform-google-modules/event-function/google"
  #  version                   = "2.5.1"
  project_id    = var.project_id
  name          = "magnus_event_function2"
  region        = "europe-west1"
  event_trigger = {
    event_type = "google.storage.object.finalize"
    resource   = "projects/${var.project_id}/buckets/${google_storage_bucket.my_upload_bucket.name}"
  }
  event_trigger_failure_policy_retry = true
  source_directory                   = local.root_dir
  runtime                            = "go120"
  entry_point                        = "HelloPubSub2"
  depends_on                         = [
    google_project_service.gcp_services
  ]
}

module "trigger-function-metatdata" {
  source        = "terraform-google-modules/event-function/google"
  #  version                   = "2.5.1"
  project_id    = var.project_id
  name          = "magnus_event_metadata"
  region        = "europe-west1"
  event_trigger = {
    event_type = "google.storage.object.metadataUpdate"
    resource   = "projects/${var.project_id}/buckets/${google_storage_bucket.my_upload_bucket.name}"
  }
  event_trigger_failure_policy_retry = true
  source_directory                   = local.root_dir
  runtime                            = "go120"
  entry_point                        = "HandleMetadataUpdated"
  depends_on                         = [
    google_project_service.gcp_services
  ]
}


resource "google_apikeys_key" "google_maps_api_key" {
  name         = "key"
  display_name = "Demo API key"
  project      = var.project_id

  restrictions {
    api_targets {
      service = "addressvalidation.googleapis.com"
      methods = ["POST"]
    }

#    browser_key_restrictions {
#      allowed_referrers = [".*"]
#    }
  }
}
