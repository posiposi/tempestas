terraform {
  required_version = "~> 1.17.0"

  backend "s3" {
    bucket       = "tempestas-tfstate-2c2fad1c"
    key          = "bootstrap/terraform.tfstate"
    region       = "ap-northeast-1"
    encrypt      = true
    use_lockfile = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.72"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.9"
    }
  }
}

provider "aws" {
  region = "ap-northeast-1"

  default_tags {
    tags = {
      Project   = "tempestas"
      ManagedBy = "terraform"
    }
  }
}
