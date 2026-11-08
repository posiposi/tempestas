output "tfstate_bucket" {
  description = "terraform/envs/dev の backend に指定する state 用バケット名"
  value       = aws_s3_bucket.tfstate.bucket
}
