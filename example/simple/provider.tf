terraform {
  backend "local" {
    path = "./default.tfstate"
  }

  required_providers {
    pipecd = {
      source  = "pipe-cd/pipecd"
      version = "0.1.0"
    }
  }

  required_version = ">= 1.4"
}

provider "pipecd" {
  # pipecd_host    = "" // optional, if not set, read from environments as PIPECD_HOST
  # pipecd_api_key = "" // optional, if not set, read from environments as PIPECD_API_KEY
}
