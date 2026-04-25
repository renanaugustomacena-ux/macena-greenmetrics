// CloudFront + WAF — edge TLS termination + bot/IP/rate protection.

variable "name_prefix"         { type = string }
variable "origin_domain"       { type = string } // ALB / ingress DNS
variable "aliases"             { type = list(string); default = ["app.greenmetrics.it"] }
variable "acm_certificate_arn" { type = string }
variable "tags"                { type = map(string); default = {} }

resource "aws_wafv2_web_acl" "this" {
  name  = "${var.name_prefix}-waf"
  scope = "CLOUDFRONT"
  default_action { allow {} }
  rule {
    name     = "AWSManagedCommonRules"
    priority = 1
    override_action { none {} }
    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "common"
      sampled_requests_enabled   = true
    }
  }
  rule {
    name     = "RateLimitPerIP"
    priority = 2
    action { block {} }
    statement {
      rate_based_statement {
        limit              = 2000
        aggregate_key_type = "IP"
      }
    }
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "ratelimit"
      sampled_requests_enabled   = true
    }
  }
  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "${var.name_prefix}-waf"
    sampled_requests_enabled   = true
  }
  tags = var.tags
}

resource "aws_cloudfront_distribution" "this" {
  enabled             = true
  is_ipv6_enabled     = true
  aliases             = var.aliases
  default_root_object = ""
  web_acl_id          = aws_wafv2_web_acl.this.arn
  comment             = "${var.name_prefix} edge"
  origin {
    domain_name = var.origin_domain
    origin_id   = "greenmetrics-origin"
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH", "DELETE"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "greenmetrics-origin"
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 86400
    compress               = true
    forwarded_values {
      query_string = true
      headers      = ["Authorization", "Origin", "Referer"]
      cookies { forward = "all" }
    }
  }
  restrictions {
    geo_restriction { restriction_type = "whitelist"; locations = ["IT", "SM", "VA", "CH", "FR", "DE", "AT", "SI"] }
  }
  viewer_certificate {
    acm_certificate_arn      = var.acm_certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
  tags = var.tags
}

output "distribution_domain" { value = aws_cloudfront_distribution.this.domain_name }
output "waf_arn"             { value = aws_wafv2_web_acl.this.arn }
