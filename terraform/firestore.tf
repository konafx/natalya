# firestore.tf

resource "google_firestore_index" "my-index" {
  project = var.project

  collection = "chatrooms"

  fields {
    field_path = "name"
    order      = "ASCENDING"
  }

  fields {
    field_path = "description"
    order      = "DESCENDING"
  }
}
