
resource "aws_iam_role" "forwarder" {
  name = "jsonl-otel-forwarder"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "forwarder" {
  name   = "jsonl-otel-forwarder"
  path   = "/"
  policy = data.aws_iam_policy_document.forwarder.json
}

resource "aws_cloudwatch_log_group" "forwarder" {
  name              = "/aws/lambda/jsonl-otel-forwarder"
  retention_in_days = 7
}

resource "aws_iam_role_policy_attachment" "forwarder" {
  role       = aws_iam_role.forwarder.name
  policy_arn = aws_iam_policy.forwarder.arn
}

data "aws_iam_policy_document" "forwarder" {
  statement {
    actions = [
      "ssm:GetParameter*",
      "ssm:DescribeParameters",
      "ssm:List*",
    ]
    resources = ["*"]
  }
  statement {
    actions = [
      "logs:GetLog*",
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["*"]
  }
}

data "archive_file" "forwarder_dummy" {
  type        = "zip"
  output_path = "${path.module}/forwarder_dummy.zip"
  source {
    content  = "forwarder_dummy"
    filename = "bootstrap"
  }
  depends_on = [
    null_resource.forwarder_dummy
  ]
}

resource "null_resource" "forwarder_dummy" {}

resource "aws_lambda_function" "forwarder" {
  lifecycle {
    ignore_changes = all
  }

  function_name = "jsonl-otel-forwarder"
  role          = aws_iam_role.forwarder.arn
  architectures = ["arm64"]
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = data.archive_file.forwarder_dummy.output_path
}

resource "aws_lambda_alias" "forwarder" {
  lifecycle {
    ignore_changes = all
  }
  name             = "current"
  function_name    = aws_lambda_function.forwarder.arn
  function_version = aws_lambda_function.forwarder.version
}

resource "aws_ssm_parameter" "header_mackerel_apikey" {
  name        = "/jsonl-otel-forwarder/FORWARDER_OTLP_HEADERS"
  description = "Mackerel API Key for Forwarder"
  type        = "SecureString"
  value       = "Mackerel-Api-Key=" + local.header_mackerel_apikey
}

resource "aws_cloudwatch_log_group" "otel_logs" {
  name              = "/jsonl-otel-forwarder/otel-logs"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_subscription_filter" "filter" {
  name            = "all"
  log_group_name  = aws_cloudwatch_log_group.otel_logs.name
  filter_pattern  = ""
  destination_arn = aws_lambda_function.forwarder.arn
  distribution    = "ByLogStream"
}

