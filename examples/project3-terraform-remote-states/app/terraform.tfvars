vpc_id1 = "vpc-443a8116aae25c7e9" # @tfvars:terraform_data.terraform_remote_state.vpc.vpc_id

vpc_id2 = "vpc-443a8116aae25c7e9" # @tfvars:terraform_output.vpc.vpc_id

# @tfvars:config:terraform_output.vpc = {
#   method = "output"  # <- this will run "terraform output"
# }
# @tfvars:config:terraform_output = {
#   method = "terraform_remote_state"  # <- this will produce terraform code with data-source and run "terraform refresh" and "terraform output"
# }
