variable "name" {}

output "welcome" {
  value = "Hello, ${var.name}!"
}