# @ tfvars:disable_annotations

name = "Anton Babenko" # @tfvars:terraform_output.core.name

score = "37" # @tfvars:terraform_output.core.score

name_as_list = [""] # @tfvars:terraform_output.core.name.to_list

love_sailing = "true" # @tfvars:terraform_output.core.love_sailing

understand_how_to_use_twitter = "false" # @tfvars:terraform_output.core.understand_how_to_use_twitter

languages = [
  "ukrainian",
  "russian",
  "english",
  "norwegian",
  "spanish",
] # @tfvars:terraform_output.core.languages
