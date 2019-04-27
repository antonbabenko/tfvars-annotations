terragrunt = {
  terraform = {
    source = "."
  }
}

################
# Static values
################

title = "This value is not going to be changed by tfvars-annotations"

#################
# Dynamic values
#################

name = "Anton Babenko" # @tfvars:terragrunt_output.core.name

score = "37" # @tfvars:terragrunt_output.core.score

love_sailing = "true" # @tfvars:terragrunt_output.core.love_sailing

understand_how_to_use_twitter = "false" # @tfvars:terragrunt_output.core.understand_how_to_use_twitter

languages = ["ukrainian", "russian", "english", "norwegian", "spanish"] # @tfvars:terragrunt_output.core.languages

list_of_properties = "" # @tfvars:terragrunt_output.core.list_of_properties

map_of_properties = "" # @tfvars:terragrunt_output.core.map_of_properties

mixed_value = "" # @tfvars:terragrunt_output.core.mixed_value

###############
# Compositions
###############

custom_map = {
  Score      = "37"            # @tfvars:terragrunt_output.core.score
  Name       = "Anton Babenko" # @tfvars:terragrunt_output.core.name
  MixedValue = ""              # @tfvars:terragrunt_output.core.mixed_value
}
