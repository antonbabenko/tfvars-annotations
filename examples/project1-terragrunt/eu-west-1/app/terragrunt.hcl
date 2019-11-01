terragrunt = {
  terraform = {
    source = "."
  }
}

dependency "core" {
  path = "../core"
}

################
# Static values
################

title = "This value is not going to be changed by tfvars-annotations"

#################
# Dynamic values
#################

name = "" # @tfvars:terragrunt_output.core.name

score = "" # @tfvars:terragrunt_output.core.score

name_as_list = [""] # @tfvars:terragrunt_output.core.name.to_list

love_sailing = "" # @tfvars:terragrunt_output.core.love_sailing

understand_how_to_use_twitter = "" # @tfvars:terragrunt_output.core.understand_how_to_use_twitter

languages = "" # @tfvars:terragrunt_output.core.languages

###############
# Compositions
###############

custom_map = {
  Score      = "" # @tfvars:terragrunt_output.core.score
  Name       = "" # @tfvars:terragrunt_output.core.name
  MixedValue = "" # @ tfvars:terragrunt_output.core.mixed_value <-- same reason as below. Maps are tricky.
}

######
# These don't work yet because there are `maps` inside of them.
######
list_of_properties = "" # @ tfvars:terragrunt_output.core.list_of_properties

map_of_properties = "" # @ tfvars:terragrunt_output.core.map_of_properties

mixed_value = "" # @ tfvars:terragrunt_output.core.mixed_value
