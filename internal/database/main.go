package database

import (
	"context"
	"expvar"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patrickmn/go-cache"
	"github.com/souravlayek/sealed-server/pkg/logger"
	"github.com/souravlayek/sealed-server/pkg/sqlc"
)

var (
	Q     *sqlc.Queries
	Pool  *pgxpool.Pool
	Cache *cache.Cache
)

func ConnectDB() *pgxpool.Pool {
	var err error
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	Pool = pool
	Q = sqlc.New(pool)

	// Register DB stats in expvar
	expvar.Publish("db_stats", expvar.Func(func() interface{} {
		stats := pool.Stat()
		return map[string]interface{}{
			"MaxConns":      stats.MaxConns(),
			"TotalConns":    stats.TotalConns(),
			"IdleConns":     stats.IdleConns(),
			"AcquiredConns": stats.AcquiredConns(),
		}
	}))

	logger.Logger.Infow("Database connected Successfully")
	// Setup Cache too
	c := cache.New(5*time.Minute, 10*time.Minute)
	Cache = c
	return pool
}
