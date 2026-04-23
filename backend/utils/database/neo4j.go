package database

import (
	"bookstore/backend/config"
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ConnectNeo4j creates a Neo4j driver and verifies connectivity.
func ConnectNeo4j(cfg config.Neo4jConfig) (neo4j.DriverWithContext, error) {
	auth := neo4j.BasicAuth(cfg.User, cfg.Password, "")

	driver, err := neo4j.NewDriverWithContext(cfg.URI, auth)
	if err != nil {
		return nil, fmt.Errorf("neo4j new driver: %w", err)
	}

	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("verify neo4j connectivity: %w", err)
	}

	return driver, nil
}
