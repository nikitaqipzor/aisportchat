package bodyscan

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/media"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

// StartMediaCleanupWorker retries committed and abandoned capture cleanup on
// startup and once per minute. Jobs remain durable after a failed attempt.
func StartMediaCleanupWorker(ctx context.Context, st store.Store, blobs media.Store) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			attempt, cancel := context.WithTimeout(ctx, 45*time.Second)
			err := st.RetryBodyScanMedia(attempt, "", func(ctx context.Context, key string) error {
				err := blobs.Delete(ctx, key)
				if errors.Is(err, media.ErrNotFound) { return nil }
				return err
			})
			if err != nil { log.Printf("body scan media cleanup retry failed: %v", err) }
			cancel()
			select { case <-ctx.Done(): return; case <-ticker.C: }
		}
	}()
}
