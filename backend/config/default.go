package config

// defaultConfig is the embedded YAML configuration loaded at startup.
// All values can be overridden via environment variables using double-underscore
// as the nested key separator (e.g. POSTGRES__HOST=myhost).
const defaultConfig = `
env: development

server:
  port: "8080"

postgres:
  host: "localhost"
  port: "5432"
  db: "bookstore"
  user: "postgres"
  password: "secret"
  sslmode: "disable"

mongo:
  uri: "mongodb://localhost:27017"
  db: "bookstore"

neo4j:
  uri: "bolt://localhost:7687"
  user: "neo4j"
  password: "password"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

jwt:
  secret: "change-me-to-a-long-random-secret"
  access_ttl: "24h"

logger:
  level: "info"
`
