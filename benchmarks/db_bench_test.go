// Package benchmarks provides performance benchmark tests for the Industrial AI Platform
// P2-001: Database Query Performance Benchmarks
// These benchmarks test database query performance using sqlmock
package benchmarks

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// Device represents a device entity for benchmark tests
type Device struct {
	ID          string
	Name        string
	Type        string
	Location    string
	Status      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BenchmarkDBConnection benchmarks database connection establishment
func BenchmarkDBConnection(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate connection pool acquisition
		mock.ExpectPing()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := db.PingContext(ctx)
		cancel()
		if err != nil {
			b.Fatalf("Ping failed: %v", err)
		}
	}
}

// BenchmarkDBQuery benchmarks basic query execution
func BenchmarkDBQuery(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("device-001", "Test Device")
		
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		_, err := db.Query("SELECT id, name FROM devices")
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkDBInsert benchmarks insert operation
func BenchmarkDBInsert(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectExec("INSERT INTO devices").
			WithArgs("device-bench-001", "Benchmark Device", "CNC", "Factory-A", "running", "", time.Now(), time.Now()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := db.Exec(
			"INSERT INTO devices (id, name, type, location, status, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			"device-bench-001", "Benchmark Device", "CNC", "Factory-A", "running", "", time.Now(), time.Now(),
		)
		if err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

// BenchmarkDBUpdate benchmarks update operation
func BenchmarkDBUpdate(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectExec("UPDATE devices").
			WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := db.Exec(
			"UPDATE devices SET name = $1, status = $2, updated_at = $3 WHERE id = $4",
			"Updated Device", "maintenance", time.Now(), "device-001",
		)
		if err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// BenchmarkDBDelete benchmarks delete operation
func BenchmarkDBDelete(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectExec("DELETE FROM devices").
			WithArgs("device-001").
			WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := db.Exec("DELETE FROM devices WHERE id = $1", "device-001")
		if err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// BenchmarkDBQueryRow benchmarks single row query
func BenchmarkDBQueryRow(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"id", "name", "type", "location", "status", "description", "created_at", "updated_at"}).
			AddRow("device-001", "Test Device", "CNC", "Factory-A", "running", "Test", time.Now(), time.Now())
		
		mock.ExpectQuery("SELECT (.+) FROM devices WHERE id").
			WithArgs("device-001").
			WillReturnRows(rows)

		var device Device
		err := db.QueryRow(
			"SELECT id, name, type, location, status, description, created_at, updated_at FROM devices WHERE id = $1",
			"device-001",
		).Scan(&device.ID, &device.Name, &device.Type, &device.Location, &device.Status, &device.Description, &device.CreatedAt, &device.UpdatedAt)
		if err != nil {
			b.Fatalf("QueryRow failed: %v", err)
		}
	}
}

// BenchmarkDBQueryWithTimeout benchmarks queries with context timeout
func BenchmarkDBQueryWithTimeout(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("1", "test")
		
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		_, err := db.QueryContext(ctx, "SELECT id, name FROM test")
		cancel()
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkDBTransaction benchmarks database transaction performance
func BenchmarkDBTransaction(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Begin()
		if err != nil {
			b.Fatalf("Begin failed: %v", err)
		}

		_, err = tx.Exec("INSERT INTO test (id) VALUES (1)")
		if err != nil {
			tx.Rollback()
			b.Fatalf("Exec failed: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			b.Fatalf("Commit failed: %v", err)
		}
	}
}

// BenchmarkDBTransactionRollback benchmarks transaction rollback
func BenchmarkDBTransactionRollback(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectRollback()

		tx, err := db.Begin()
		if err != nil {
			b.Fatalf("Begin failed: %v", err)
		}

		err = tx.Rollback()
		if err != nil {
			b.Fatalf("Rollback failed: %v", err)
		}
	}
}

// BenchmarkDBConcurrentQueries benchmarks concurrent database queries
func BenchmarkDBConcurrentQueries(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rows := sqlmock.NewRows([]string{"id", "name"}).
				AddRow("1", "test")
			
			mock.ExpectQuery("SELECT").
				WillReturnRows(rows)

			_, err := db.Query("SELECT id, name FROM devices")
			if err != nil {
				continue
			}
		}
	})
}

// BenchmarkDBBulkInsert benchmarks bulk insert performance
func BenchmarkDBBulkInsert(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Prepare 100 devices for bulk insert
	devices := make([]Device, 100)
	for i := 0; i < 100; i++ {
		devices[i] = Device{
			ID:        fmt.Sprintf("device-bulk-%d", i),
			Name:      fmt.Sprintf("Bulk Device %d", i),
			Type:      "CNC",
			Location:  "Factory-A",
			Status:    "running",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		for j := 0; j < 100; j++ {
			mock.ExpectExec("INSERT").
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectCommit()

		tx, _ := db.Begin()
		for _, device := range devices {
			_, err := tx.Exec(
				"INSERT INTO devices (id, name, type, location, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
				device.ID, device.Name, device.Type, device.Location, device.Status, device.CreatedAt, device.UpdatedAt,
			)
			if err != nil {
				tx.Rollback()
				break
			}
		}
		tx.Commit()
	}
}

// BenchmarkDBScanPerformance benchmarks row scanning performance
func BenchmarkDBScanPerformance(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Create rows with many columns
	columns := []string{"id", "name", "type", "location", "status", "description", 
		"created_at", "updated_at", "value1", "value2", "value3", "value4", "value5", "value6"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows(columns)
		for j := 0; j < 50; j++ {
			rows.AddRow(j, "name", "type", "loc", "status", "desc", 
				time.Now(), time.Now(), 1.0, 2.0, 3.0, 4.0, 5.0, 6.0)
		}
		
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		result, _ := db.Query("SELECT * FROM devices")
		for result.Next() {
			var id int
			var name, typ, loc, status, desc string
			var created, updated time.Time
			var v1, v2, v3, v4, v5, v6 float64
			_ = result.Scan(&id, &name, &typ, &loc, &status, &desc, 
				&created, &updated, &v1, &v2, &v3, &v4, &v5, &v6)
		}
		result.Close()
	}
}

// BenchmarkDBPreparedStatement benchmarks prepared statement performance
func BenchmarkDBPreparedStatement(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectPrepare("SELECT")
		mock.ExpectQuery("SELECT").
			WithArgs("device-001").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("device-001"))

		stmt, err := db.Prepare("SELECT id FROM devices WHERE id = $1")
		if err != nil {
			b.Fatalf("Prepare failed: %v", err)
		}

		_, err = stmt.Query("device-001")
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
		stmt.Close()
	}
}

// BenchmarkDBConnectionPool benchmarks connection pool operations
func BenchmarkDBConnectionPool(b *testing.B) {
	var mu sync.Mutex
	pool := make([]*sql.DB, 0)
	poolSize := 10

	// Simulate connection pool
	for i := 0; i < poolSize; i++ {
		db, _, _ := sqlmock.New()
		pool = append(pool, db)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			if len(pool) > 0 {
				db := pool[0]
				pool = pool[1:]
				mu.Unlock()
				
				// Use connection
				_ = db.Ping()
				
				// Return to pool
				mu.Lock()
				pool = append(pool, db)
				mu.Unlock()
			} else {
				mu.Unlock()
			}
		}
	})

	// Cleanup
	for _, db := range pool {
		db.Close()
	}
}

// BenchmarkDBQueryBuilder benchmarks SQL query building
func BenchmarkDBQueryBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := "SELECT id, name, type, location, status FROM devices WHERE "
		conditions := []string{}
		
		if i%2 == 0 {
			conditions = append(conditions, "status = 'running'")
		}
		if i%3 == 0 {
			conditions = append(conditions, "type = 'CNC'")
		}
		if i%5 == 0 {
			conditions = append(conditions, "location LIKE 'Factory-%'")
		}
		
		for j, cond := range conditions {
			if j > 0 {
				query += " AND "
			}
			query += cond
		}
		
		query += " ORDER BY created_at DESC LIMIT 100"
		_ = query
	}
}

// BenchmarkDBResultSetIteration benchmarks result set iteration
func BenchmarkDBResultSetIteration(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"id", "value"})
		for j := 0; j < 1000; j++ {
			rows.AddRow(j, j*1.5)
		}
		
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		result, _ := db.Query("SELECT id, value FROM telemetry")
		count := 0
		for result.Next() {
			count++
		}
		result.Close()
	}
}

// BenchmarkDBCountQuery benchmarks count query
func BenchmarkDBCountQuery(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT COUNT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1000))

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&count)
		if err != nil {
			b.Fatalf("Count failed: %v", err)
		}
	}
}

// BenchmarkDBAggregation benchmarks aggregation queries
func BenchmarkDBAggregation(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT COUNT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10000))
		mock.ExpectQuery("SELECT AVG").
			WillReturnRows(sqlmock.NewRows([]string{"avg"}).AddRow(75.5))
		mock.ExpectQuery("SELECT MAX").
			WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(120.0))

		// Simulate aggregation
		_, _ = db.Query("SELECT COUNT(*) FROM telemetry")
		_, _ = db.Query("SELECT AVG(temperature) FROM telemetry")
		_, _ = db.Query("SELECT MAX(temperature) FROM telemetry")
	}
}

// BenchmarkDBNullHandling benchmarks handling NULL values
func BenchmarkDBNullHandling(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"id", "value", "optional"}).
			AddRow(1, 75.5, nil).
			AddRow(2, 80.0, "data")
		
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		result, _ := db.Query("SELECT id, value, optional FROM telemetry")
		for result.Next() {
			var id int
			var value float64
			var optional sql.NullString
			_ = result.Scan(&id, &value, &optional)
		}
		result.Close()
	}
}

// BenchmarkDBPaginatedQuery benchmarks paginated queries
func BenchmarkDBPaginatedQuery(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mock count query
		mock.ExpectQuery("SELECT COUNT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1000))

		// Mock paginated query
		rows := sqlmock.NewRows([]string{"id", "name", "type"})
		for j := 0; j < 100; j++ {
			rows.AddRow(fmt.Sprintf("device-%d", j), fmt.Sprintf("Device %d", j), "CNC")
		}
		
		mock.ExpectQuery("SELECT (.+) FROM devices ORDER BY").
			WithArgs(100, 0).
			WillReturnRows(rows)

		_, _ = db.Query("SELECT id, name, type FROM devices ORDER BY created_at DESC LIMIT $1 OFFSET $2", 100, 0)
	}
}

// BenchmarkDBUpsert benchmarks upsert (INSERT ON CONFLICT UPDATE) operation
func BenchmarkDBUpsert(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectExec("INSERT INTO devices").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := db.Exec(
			`INSERT INTO devices (id, name, type, location, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at`,
			"device-001", "Upsert Device", "CNC", "Factory-A", "running", time.Now(), time.Now(),
		)
		if err != nil {
			b.Fatalf("Upsert failed: %v", err)
		}
	}
}

// BenchmarkDBIndexQuery benchmarks indexed query
func BenchmarkDBIndexQuery(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT (.+) FROM devices WHERE").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("device-001"))

		// Simulate index lookup
		_, err := db.Query("SELECT id FROM devices WHERE device_id = $1", "device-001")
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkDBMultipleScans benchmarks multiple scan operations
func BenchmarkDBMultipleScans(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "one").
			AddRow(2, "two").
			AddRow(3, "three")
		
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		result, _ := db.Query("SELECT id, name FROM devices")
		var ids []int
		var names []string
		for result.Next() {
			var id int
			var name string
			result.Scan(&id, &name)
			ids = append(ids, id)
			names = append(names, name)
		}
		result.Close()
	}
}

// BenchmarkDBBatchUpdate benchmarks batch update operations
func BenchmarkDBBatchUpdate(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		for j := 0; j < 50; j++ {
			mock.ExpectExec("UPDATE").
				WillReturnResult(sqlmock.NewResult(0, 1))
		}
		mock.ExpectCommit()

		tx, _ := db.Begin()
		for j := 0; j < 50; j++ {
			tx.Exec("UPDATE devices SET status = $1 WHERE id = $2", "updated", fmt.Sprintf("device-%d", j))
		}
		tx.Commit()
	}
}

// BenchmarkDBJoinQuery benchmarks join queries
func BenchmarkDBJoinQuery(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := sqlmock.NewRows([]string{"device_id", "device_name", "telemetry_id", "value"})
		for j := 0; j < 10; j++ {
			rows.AddRow(fmt.Sprintf("device-%d", j), fmt.Sprintf("Device %d", j), j, j*10.0)
		}
		
		mock.ExpectQuery("SELECT (.+) JOIN").
			WillReturnRows(rows)

		_, err := db.Query(`
			SELECT d.id, d.name, t.id, t.value 
			FROM devices d 
			JOIN telemetry t ON d.id = t.device_id 
			WHERE d.status = $1`, "running")
		if err != nil {
			b.Fatalf("Join failed: %v", err)
		}
	}
}