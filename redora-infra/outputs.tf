output "cloudsql_connection_name" {
  value = google_sql_database_instance.postgres.connection_name
}

output "redis_host" {
  value = google_redis_instance.memorystore.host
}

output "project_id" {
  value = var.project_id
}