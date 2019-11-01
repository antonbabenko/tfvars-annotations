data "terraform_remote_state" "vpc" {
  backend = "s3"

  config {
    bucket = "my-tfstates"
    key = "vpc/terraform.tfstate"
  }
}

output "vpc_id" {
  value = "${data.terraform_remote_state.vpc.vpc_id}"
}