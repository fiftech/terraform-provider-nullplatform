data "nullplatform_application" "app" {
  id = var.null_application_id
}

resource "nullplatform_service" "redis_cache_test" {
  name             =  "redis-cache"
  specification_id = "4a4f6955-5ae0-40dc-a1de-e15e5cf41abb"
  entity_nrn       = data.nullplatform_application.app.nrn
  linkable_to      = [data.nullplatform_application.app.nrn]
  dimensions = {}
  selectors = {
    imported = false,
  }
  attributes = {}
}

data "nullplatform_service" "redis" {
  id = nullplatform_service.redis_cache_test.id
}

resource "nullplatform_link" "link_redis" {
  name             = "link_from_terraform_2"
  status           = "active"
  service_id       = data.nullplatform_service.redis.id
  specification_id = "66919464-05e6-4d78-bb8c-902c57881ddd"
  entity_nrn       = data.nullplatform_application.app.nrn
  linkable_to      = [data.nullplatform_application.app.nrn]
  selectors = {
    imported = false,
  }
  dimensions = {
    environment = "development",
    country     = "argentina",
  }
  attributes = {}
}