output "name" {
  value = "Anton Babenko"
}

output "score" {
  value = "37"
}

output "love_sailing" {
  value = "true"
}

output "understand_how_to_use_twitter" {
  value = "false"
}

output "languages" {
  value = ["ukrainian", "russian", "english", "norwegian", "spanish"]
}

output "map_of_properties" {
  value = {
    Name        = "Anton Babenko"
    Age         = 34
    LoveSailing = true
  }
}

output "list_of_properties" {
  value = [
    {
      Name                      = "Anton Babenko"
      Age                       = 34
      LoveSailing               = true
      UnderstandHowToUseTwitter = false
    },
    {
      Name        = "Kapitoshka"
      Age         = 123
      LoveSailing = false
    },
  ]
}

output "mixed_value" {
  value = [
    "This is just a string",
    [
      {
        Name                      = "Anton Babenko"
        Age                       = 34
        LoveSailing               = true
        UnderstandHowToUseTwitter = false
      },
    ],
    {
      Github  = "antonbabenko"
      Twitter = "antonbabenko"
    },
  ]
}

// Failing values ("=" inside values):
output "failing_values_list" {
  value = ["ukrainian = 100", "english", "unknown"]
}

output "failing_values_map" {
  value = {
    Name = "Anton Babenko"
    Age  = 34
  }
}
