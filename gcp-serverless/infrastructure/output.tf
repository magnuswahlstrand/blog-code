
output "adress_validation_api_key" {
    value = google_apikeys_key.google_maps_api_key.key_string
    sensitive = true
}