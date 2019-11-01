# Update values in terraform.tfvars using annotations

### This is still WIP, full of bugs, missing core features and docs.

[Terraform](https://www.terraform.io/) is awesome!
 
Terraform 0.11 and 0.12 supports only static (known, fixed, already computed) values in `tfvars` files. There is no way to use Terraform functions, or data-sources inside `tfvars` files in Terraform to update values.

Using Terraform modules (eg, terraform-aws-modules) developers are not restricted on specific ways how module arguments can be provided (eg, from modules, remote states, other data-sources, command line arguments, tfvars files, etc). Using tfvars files is often a very good idea, because it separates configurations and values! And this is exactly where `tfvars-annotations` tool helps.

While working on [modules.tf](https://github.com/antonbabenko/modules.tf-lambda) (a tool which converts visual diagrams created with [Cloudcraft.co](https://cloudcraft.co/) into Terraform configurations), I had a need to generate code which would chain invocations of [Terraform AWS modules](https://github.com/terraform-aws-modules) and pass arguments between them without requiring any extra Terraform code as a glue. [Terragrunt](https://github.com/gruntwork-io/terragrunt) is a great fit for this, it allows to reduce amount of Terraform configurations by reusing Terraform modules and providing arguments as values in `tfvars` files.

Some languages I know have concepts like annotations and decorators, so at first I made a [shell script](https://github.com/antonbabenko/modules.tf-lambda/blob/v1.2.0/templates/terragrunt-common-layer/common/scripts/update_dynamic_values_in_tfvars.sh) which replaced values in `tfvars` based on annotations and was called by Terragrunt hooks. `tfvars-annotations` shares the same goal and it has no external dependencies (except `terraform` or `terragrunt`).


## Use cases

1. modules.tf and Terragrunt (recommended)
1. Terraform (not implemented yet)
1. General example - AMI ID, AWS region

## Features

1. Supported annotations:
  - [x] terragrunt_output:
     - `@tfvars:terragrunt_output.vpc.vpc_id`
     - `@tfvars:terragrunt_output.security-group.this_security_group_id`
  - [x] terraform_output
     - `@tfvars:terraform_output.vpc.vpc_id`
     - `@tfvars:terraform_output.security-group.this_security_group_id`
  - [ ] data-sources generic
     - `@tfvars:terragrunt_data.aws_region.zone_id` ?
     - `@tfvars:terraform_data.aws_region.zone_id` ?

## General form of the annotations:

`@tfvars:TYPE.DIRECTORY.ATTR[.RETURN_TYPE]`

`TYPE` - Type of the fetcher:
1. terragrunt_output - Get value from `terragrunt output <ATTR>`
1. terraform_output - Get value from `terraform output <ATTR>`
1. terraform_data and terragrunt_data - Get attribute from data-source (eg, `aws_caller_identity.account_id` to ret)

`DIRECTORY` - Working directory where `TYPE` commands should be run. It can be specified in short form if it is a name of directory on the same level as parent directory.
Long form should be used to specify directories on the same level as well as from parent (no max level).

Examples:
1. `[./core]` and `[core]` are identical and refer to `core` subfolder, `[../../../core]` - several folders above
1. `[../core]` and `core` have identical meanings and refer to `core` in the parent directory
1. `[..]` - parent directory
1. `[.]` - current directory

`ATTR` - Name of output to return or name of attribute if `TYPE` is data-source (terraform_data and terragrunt_data).

`RETURN_TYPE` - (Optional) Specify desired type to return. `to_list` can be used to wrap original value with `[]` to make it as a list.

## Examples of various annotations

1. `@tfvars:terragrunt_output.vpc.vpc_id`
1. `@tfvars:terragrunt_output.[../../vpc].vpc_id`
1. `@tfvars:terraform_output.[./vpc].vpc_id.to_list`
1. `@tfvars:terraform_output.[../us-east-1/acm].this_acm_arn`
1. `@tfvars:terragrunt_data.aws_region.zone_id`
1. `@tfvars:terragrunt_data.aws_caller_identity.account_id.to_list`

## How to use

```
Usage: tfvars-annotations [DIR]
  -debug
        enable debug logging
  -version
        print version information and exit

[DIR] (optional) - If not specified current directory will be used.
```

Run `tfvars-annotations` before `terraform plan, apply, refresh`.

It will process tfvars file in the current directory and set updated values.

E.g.:

    $ tfvars-annotations examples/project1-terragrunt/eu-west-1/app
    $ terraform plan

## Annotations

## How to disable processing entirely

Put `@tfvars:disable_annotations` anywhere in the `terraform.tfvars` to not process the file.

## Examples

See `examples` for some basics.

## To-do

1. Get values from other sources:
 - data sources generic (with missing values to use the default value)
 - aws_account_id or aws_region data sources
 - aws_account_id by alias (reference by alias is friendlier)
2. terragrunt_outputs from stacks:
 - in any folder
 - in current region
3. cache values unless stack is changed/updated
4. functions (limit(2), to_list)
5. rewrite in go (invoke like this => update_dynamic_values_in_tfvars ${get_parent_tfvars_dir()}/${path_relative_to_include()})
6. make it much faster, less verbose
7. add dry-run flag
9. Add -detailed-exitcode to know if files were changed
10. Support terraform.tfvars and terragrunt.hcl or custom list of files
8. Proposed syntax:

 - `@tfvars:terragrunt_output.security-group_5.this_security_group_id.to_list`

 - `@tfvars:terragrunt_output.["eu-west-1/security-group_5"].this_security_group_id.to_list`

 - `@tfvars:terragrunt_output.["global/route53-zones"].zone_id`

 - `@tfvars:terragrunt_data.aws_region.zone_id`

 - `@tfvars:terragrunt_data.aws_region[{current=true}].zone_id`
// 9. Describe and compare tfvars-annotations vs using remote_state

## Bugs

1. Add support for `maps` (and lists of maps). Strange bugs with rendering comments in wrong places.

## Installation

On OSX, install it with Homebrew (not enough github stars to get it to the official repo):

```
brew install -s HomebrewFormula/tfvars-annotations.rb
```

Alternatively, you can download a [release](https://github.com/antonbabenko/tfvars-annotations/releases) suitable for your platform and unzip it. Make sure the `tfvars-annotations` binary is executable, and you're ready to go.

You can also install it like this:

```
go get github.com/antonbabenko/tfvars-annotations
```

Or run it from source:

```
go run . -debug examples/project1-terragrunt/eu-west-1/app
go run . examples/project1-terragrunt/eu-west-1/app
```

## Release

1. Make GitHub Release: `hub release create v0.0.3`. Then Github Actions will build binaries and attach them to Github release.
2. Update Homebrew version in `HomebrewFormula/tfvars-annotations.rb`. Install locally - `brew install -s HomebrewFormula/tfvars-annotations.rb`

## Authors

This project is created and maintained by [Anton Babenko](https://github.com/antonbabenko) with the help from [different contributors](https://github.com/antonbabenko/tfvars-annotations/graphs/contributors).

[![@antonbabenko](https://img.shields.io/twitter/follow/antonbabenko.svg?style=social&label=Follow%20@antonbabenko%20on%20Twitter)](https://twitter.com/antonbabenko)


## License

This work is licensed under MIT License. See LICENSE for full details.

Copyright (c) 2019 Anton Babenko
