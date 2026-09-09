package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"access_expires_at"`
}

type Service struct {
	store  store.Store
	tokens *TokenManager
}

func NewService(s store.Store, t *TokenManager) *Service { return &Service{store: s, tokens: t} }

func (s *Service) Register(ctx context.Context, email, password string) (store.User, Tokens, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") {
		return store.User{}, Tokens{}, errors.New("valid email is required")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return store.User{}, Tokens{}, err
	}
	user, err := s.store.CreateUser(ctx, email, hash)
	if err != nil {
		return store.User{}, Tokens{}, err
	}
	tokens, err := s.issueTokens(ctx, user.ID)
	return user, tokens, err
}

func (s *Service) Login(ctx context.Context, email, password string) (store.User, Tokens, error) {
	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil || !VerifyPassword(user.PasswordHash, password) {
		return store.User{}, Tokens{}, ErrInvalidCredentials
	}
	tokens, err := s.issueTokens(ctx, user.ID)
	return user, tokens, err
}

func (s *Service) Refresh(ctx context.Context, refreshRaw string) (Tokens, error) {
	hash := HashRefreshToken(refreshRaw)
	sess, err := s.store.GetRefreshSession(ctx, hash)
	if err != nil || sess.RevokedAt != nil || time.Now().UTC().After(sess.ExpiresAt) {
		return Tokens{}, errors.New("invalid refresh token")
	}
	_ = s.store.RevokeRefreshSession(ctx, hash)
	return s.issueTokens(ctx, sess.UserID)
}

func (s *Service) Logout(ctx context.Context, refreshRaw string) error {
	if refreshRaw == "" {
		return nil
	}
	err := s.store.RevokeRefreshSession(ctx, HashRefreshToken(refreshRaw))
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	return err
}

func (s *Service) issueTokens(ctx context.Context, userID string) (Tokens, error) {
	access, exp, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return Tokens{}, err
	}
	refresh, hash, refreshExp := s.tokens.NewRefreshToken()
	if err := s.store.SaveRefreshSession(ctx, store.RefreshSession{TokenHash: hash, UserID: userID, ExpiresAt: refreshExp}); err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: refresh, ExpiresAt: exp}, nil
}
