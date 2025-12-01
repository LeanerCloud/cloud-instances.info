resource "aws_iam_user" "pricing" {
  force_destroy = false
  name          = "pricing"
  path          = "/"
}

resource "aws_iam_user_policy_attachment" "pricing" {
  user       = aws_iam_user.pricing.name
  policy_arn = aws_iam_policy.ec2_pricing.arn
}

resource "aws_iam_policy" "ec2_pricing" {
  name        = "ec2_pricing"
  description = "allow access to ec2 instance types and pricing information"
  path        = "/"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ec2:DescribeInstanceTypeOfferings",
          "ec2:DescribeInstanceTypes",
          "ec2:DescribeRegions",
          "ec2:DescribeSpotPriceHistory",
          "elasticache:DescribeEngineDefaultParameters",
          "pricing:GetProducts",
        ],
        Resource = "*"
      }
    ]
  })
  tags = {
    Terraformed = "true"
  }
}


resource "aws_iam_policy" "terraform_state_access" {
  name        = "terraform_state_access"
  description = "Allow access to S3 bucket for Terraform state storage"
  path        = "/"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ],
        Resource = [
          aws_s3_bucket.terraform_state.arn,
          "${aws_s3_bucket.terraform_state.arn}/*"
        ]
      }
    ]
  })
  tags = {
    Terraformed = "true"
  }
}

resource "aws_iam_user_policy_attachment" "pricing_terraform_state" {
  user       = aws_iam_user.pricing.name
  policy_arn = aws_iam_policy.terraform_state_access.arn
}
