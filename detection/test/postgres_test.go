package test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/rogueprox/liquidgold/detection"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDetectPostgres(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15.3-alpine"),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	ok, err := detection.IsPostgresql(
		ctx,
		"localhost",
		mappedPort.Int(),
	)
	require.NoError(t, err)

	require.True(t, ok)

	notOk, err := detection.IsPostgresql(ctx, "localhost", 22)
	require.NoError(t, err)

	require.False(t, notOk)

}

func BenchmarkTestIsPostgres(b *testing.B) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15.3-alpine"),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			b.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	require.NoError(b, err)

	mappedPortInt := mappedPort.Int()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = detection.IsPostgresql(
			ctx,
			"localhost",
			mappedPortInt,
		)
	}
	b.StopTimer()
}

func TestDetectSlow(t *testing.T) {
	ln, err := net.Listen("tcp4", ":0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port

	timeout := time.Second * 5

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("Listening err: %v", err)
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(timeout * 2):
			}
			_, _ = conn.Write([]byte{1, 2, 4})
			conn.Close()
		}
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ok, err := detection.IsPostgresql(timeoutCtx, "localhost", port)
	require.False(t, ok)
}


