variable "project" {
  description = "gcp project"
  type = string
  default = null
}

variable "region" {
  description = "region to use the module"
  type = string
  default = "us-east1"
}

variable "zone" {
  description = "zone"
  type =string
  default = "us-east1-a"
}
