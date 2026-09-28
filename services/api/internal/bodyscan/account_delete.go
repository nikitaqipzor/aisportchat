package bodyscan

import "context"

// DeleteUserMedia removes all body scan files, including files left behind by
// an interrupted photo replacement. Caller must still hold the account store
// deletion lock so associated metadata cannot be removed on cleanup failure.
func (s *Service) DeleteUserMedia(ctx context.Context, userID string) error {
	return s.media.DeleteUserMedia(ctx, userID)
}
