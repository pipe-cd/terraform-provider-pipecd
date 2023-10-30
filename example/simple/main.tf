data "pipecd_piped" "main" {
  id = var.piped_id
}

resource "pipecd_application" "main" {
  kind              = "CLOUDRUN"
  name              = "example-application"
  description       = "This is the simple application by ${data.pipecd_piped.main.name}"
  platform_provider = "cloudrun-inproject"
  piped_id          = data.pipecd_piped.main.id
  git = {
    repository_id = "examples"
    path          = "cloudrun/simple"
    filename      = "app.pipecd.yaml"
  }
}
