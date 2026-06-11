package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

const (
	PG_CONN_STR = "postgres://postgres:password@localhost:5432/testdb?sslmode=disable"
	PEBBLE_DIR  = "./pebble_test_db"
)

func generateMessages(count int, bN int, iter int) []Message {
	msgs := make([]Message, count)
	for i := range count {
		msgs[i] = Message{
			ID:        fmt.Sprintf("msg-%d-%d-%d", bN, iter, i),
			Content:   "This is a sample message content used for benchmarking storage performance.",
			Timestamp: time.Now().UnixNano(),
		}
	}
	return msgs
}

// ==========================================
// Pebble 基准测试
// ==========================================
func benchmarkPebble(b *testing.B, batchSize int) {
	db, err := pebble.Open(PEBBLE_DIR, &pebble.Options{})
	if err != nil {
		b.Fatalf("Failed to open pebble: %v", err)
	}
	defer os.RemoveAll(PEBBLE_DIR)
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msgs := generateMessages(batchSize, b.N, i)

		batch := db.NewBatch()
		for _, msg := range msgs {
			key := []byte(msg.ID)
			val, _ := json.Marshal(msg)
			if err := batch.Set(key, val, pebble.NoSync); err != nil {
				b.Fatal(err)
			}
		}
		if err := batch.Commit(pebble.Sync); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPebble_1K(b *testing.B)  { benchmarkPebble(b, 1000) }
func BenchmarkPebble_5K(b *testing.B)  { benchmarkPebble(b, 5000) }
func BenchmarkPebble_10K(b *testing.B) { benchmarkPebble(b, 10000) }

// ==========================================
// PostgreSQL + pgx 基准测试
// ==========================================
func benchmarkPGX(b *testing.B, batchSize int) {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(PG_CONN_STR)
	if err != nil {
		b.Fatalf("Parse config failed: %v", err)
	}
	config.MaxConns = 50

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		b.Fatalf("Connect to PG failed: %v", err)
	}
	defer pool.Close()

	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS messages;")
	_, err = pool.Exec(ctx, `
		CREATE TABLE messages (
			id VARCHAR(64) PRIMARY KEY,
			content TEXT,
			timestamp BIGINT
		);
	`)
	if err != nil {
		b.Fatalf("Create table failed: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msgs := generateMessages(batchSize, b.N, i)
		rows := make([][]any, len(msgs))
		for idx, msg := range msgs {
			rows[idx] = []any{msg.ID, msg.Content, msg.Timestamp}
		}
		_, err := pool.CopyFrom(
			ctx,
			pgx.Identifier{"messages"},
			[]string{"id", "content", "timestamp"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPGX_1K(b *testing.B)  { benchmarkPGX(b, 1000) }
func BenchmarkPGX_5K(b *testing.B)  { benchmarkPGX(b, 5000) }
func BenchmarkPGX_10K(b *testing.B) { benchmarkPGX(b, 10000) }
