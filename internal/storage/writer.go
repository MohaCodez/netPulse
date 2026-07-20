package storage

import (
	"log"
	"sync"
	"time"
)

// Writer is a single-goroutine database writer that batches inserts.
// All writes go through this to avoid SQLite contention.
type Writer struct {
	db      *DB
	queue   chan func()
	done    chan struct{}
	wg      sync.WaitGroup
}

// NewWriter creates a batching database writer.
func NewWriter(db *DB, bufferSize int) *Writer {
	w := &Writer{
		db:    db,
		queue: make(chan func(), bufferSize),
		done:  make(chan struct{}),
	}
	w.wg.Add(1)
	go w.loop()
	return w
}

// Stop gracefully shuts down the writer, flushing pending writes.
func (w *Writer) Stop() {
	close(w.done)
	w.wg.Wait()
}

// Enqueue adds a write operation to the queue. Retries once if queue is full.
func (w *Writer) Enqueue(fn func()) {
	select {
	case w.queue <- fn:
	default:
		// Queue full — wait briefly then retry once
		time.Sleep(5 * time.Millisecond)
		select {
		case w.queue <- fn:
		default:
			log.Printf("[writer] write dropped: queue full (%d items)", len(w.queue))
		}
	}
}

func (w *Writer) loop() {
	defer w.wg.Done()

	batch := make([]func(), 0, 32)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			// Drain remaining
			w.flush(batch)
			for {
				select {
				case fn := <-w.queue:
					fn()
				default:
					return
				}
			}
		case fn := <-w.queue:
			batch = append(batch, fn)
			// Keep collecting if more are ready
			for len(batch) < 32 {
				select {
				case fn := <-w.queue:
					batch = append(batch, fn)
				default:
					goto flush
				}
			}
		flush:
			w.flush(batch)
			batch = batch[:0]
		case <-ticker.C:
			if len(batch) > 0 {
				w.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

func (w *Writer) flush(batch []func()) {
	if len(batch) == 0 {
		return
	}

	tx, err := w.db.conn.Begin()
	if err != nil {
		log.Printf("[writer] begin tx error: %v", err)
		// Fall back to individual writes
		for _, fn := range batch {
			fn()
		}
		return
	}

	for _, fn := range batch {
		fn()
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[writer] commit error: %v", err)
		tx.Rollback()
	}
}
