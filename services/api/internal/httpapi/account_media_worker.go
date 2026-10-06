package httpapi

import (
	"context"
	"log"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/media"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

// StartMediaCleanupWorker retries durable deletion jobs once at startup and
// every minute. The database outbox survives crashes and transient disk errors.
func StartMediaCleanupWorker(ctx context.Context, st store.Store, blobs media.Store) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			attemptCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			if err := st.RetryPendingMedia(attemptCtx, blobs.DeleteUserMedia); err != nil {
				log.Printf("account media cleanup retry failed: %v", err)
			}
			cancel()
			select {
			case <-ctx.Done(): return
			case <-ticker.C:
			}
		}
	}()
}
