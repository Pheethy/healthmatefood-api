package database

import (
	"context"
	"healthmatefood-api/config"

	"github.com/Pheethy/psql"
	"github.com/Pheethy/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/opentracing/opentracing-go"
)

const (
	PGX  = "pgx"
	SQLX = "sqlx"
)

func DBConnect(ctx context.Context, cfg config.IDbConfig, tracing opentracing.Tracer) *sqlx.DB {
	/* connect */
	psqlClient := getPostgresClient(cfg.Url(), tracing)
	db := psqlClient.GetClient()
	db.SetMaxOpenConns(cfg.MaxConns())

	return db
}

func getPostgresClient(conn string, tracing opentracing.Tracer) *psql.Client {
	client, err := psql.NewPsqlWithTracingConnection(conn, tracing)
	if err != nil {
		panic(err)
	}
	return client
}
