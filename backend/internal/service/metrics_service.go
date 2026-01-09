package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	repository "github.com/moh-sso-dashboard/internal/repository/metrics"
)

type MetricsService struct {
	repo repository.MetricsRepository
}

func NewMetricsService(repo repository.MetricsRepository) *MetricsService {
	return &MetricsService{repo: repo}
}

func (s *MetricsService) CountUsers(ctx context.Context) (int64, error) {
	return s.repo.CountUsers(ctx)
}

func (s *MetricsService) CountDisabledUsers(ctx context.Context) (int64, error) {
	return s.repo.CountDisabledUsers(ctx)
}

func (s *MetricsService) CountActiveUsers(ctx context.Context) (int64, error) {
	return s.repo.CountActiveUsers(ctx)
}

func (s *MetricsService) CountActiveUsersInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return s.repo.CountActiveUsersInRange(ctx, start, end)
}

func (s *MetricsService) ActiveUsersToday(ctx context.Context) (int64, error) {
	return s.repo.ActiveUsersToday(ctx)
}

func (s *MetricsService) ActiveUsersThisWeek(ctx context.Context) (int64, error) {
	return s.repo.ActiveUsersThisWeek(ctx)
}

func (s *MetricsService) LoginTrend(ctx context.Context) ([]db.LoginTrendRow, error) {
	return s.repo.LoginTrend(ctx)
}

func (s *MetricsService) LoginTrendByDay(ctx context.Context, start, end time.Time) ([]db.LoginTrendByDayRow, error) {
	return s.repo.LoginTrendByDay(ctx, start, end)
}

func (s *MetricsService) LoginSuccessFailure(ctx context.Context, start, end time.Time) (db.LoginSuccessFailureInRangeRow, error) {
	return s.repo.LoginSuccessFailureInRange(ctx, start, end)
}

func (s *MetricsService) TotalLoginsInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return s.repo.TotalLoginsInRange(ctx, start, end)
}

func (s *MetricsService) ApproximateActiveSessions(ctx context.Context) (int64, error) {
	return s.repo.ApproximateActiveSessions(ctx)
}

func (s *MetricsService) CountFailedLogins(ctx context.Context) (int64, error) {
	return s.repo.CountFailedLogins(ctx)
}

func (s *MetricsService) CountFailedLoginsInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return s.repo.CountFailedLoginsInRange(ctx, start, end)
}

func (s *MetricsService) CountPasswordResetsInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return s.repo.CountPasswordResetsInRange(ctx, start, end)
}

func (s *MetricsService) SuspiciousLogins(ctx context.Context, start, end time.Time, homeCountry string) ([]db.SuspiciousLoginsInRangeRow, error) {
	return s.repo.SuspiciousLoginsInRange(ctx, start, end, homeCountry)
}

func (s *MetricsService) CountClients(ctx context.Context) (int64, error) {
	return s.repo.CountClients(ctx)
}

func (s *MetricsService) CountEnabledClients(ctx context.Context) (int64, error) {
	return s.repo.CountEnabledClients(ctx)
}

func (s *MetricsService) MostActiveClients(ctx context.Context) ([]db.MostActiveClientsRow, error) {
	return s.repo.MostActiveClients(ctx)
}

func (s *MetricsService) MostAccessedClients(ctx context.Context, start, end time.Time, limit int32) ([]db.MostAccessedClientsRow, error) {
	return s.repo.MostAccessedClients(ctx, start, end, limit)
}

func (s *MetricsService) LoginCountForClient(ctx context.Context, clientID string, start, end time.Time) (int64, error) {
	return s.repo.LoginCountForClientInRange(ctx, clientID, start, end)
}

func (s *MetricsService) ActiveUsersPerClientToday(ctx context.Context) ([]db.ActiveUsersPerClientTodayRow, error) {
	return s.repo.ActiveUsersPerClientToday(ctx)
}

func (s *MetricsService) NewClientsInRange(ctx context.Context, start, end time.Time) ([]db.Client, error) {
	return s.repo.NewClientsInRange(ctx, start, end)
}

func (s *MetricsService) RecentlyCreatedClients(ctx context.Context, limit int32) ([]db.Client, error) {
	return s.repo.RecentlyCreatedClients(ctx, limit)
}

func (s *MetricsService) NewUsersInRange(ctx context.Context, start, end time.Time) ([]db.User, error) {
	return s.repo.NewUsersInRange(ctx, start, end)
}

func (s *MetricsService) NewUsersTrend(ctx context.Context, start, end time.Time) ([]db.NewUsersTrendRow, error) {
	return s.repo.NewUsersTrend(ctx, start, end)
}

func (s *MetricsService) CountNewUsersToday(ctx context.Context) (int64, error) {
	return s.repo.CountNewUsersToday(ctx)
}

func (s *MetricsService) CountNewUsersThisWeek(ctx context.Context) (int64, error) {
	return s.repo.CountNewUsersThisWeek(ctx)
}

func (s *MetricsService) NeverLoggedInUsers(ctx context.Context) ([]db.User, error) {
	return s.repo.NeverLoggedInUsers(ctx)
}

func (s *MetricsService) LastLoginForUser(ctx context.Context, userID uuid.UUID) (sql.NullTime, error) {
	return s.repo.LastLoginForUser(ctx, userID)
}

func (s *MetricsService) LastLoginForAllUsers(ctx context.Context) ([]db.LastLoginForAllUsersRow, error) {
	return s.repo.LastLoginForAllUsers(ctx)
}

func (s *MetricsService) LoginCountPerUser(ctx context.Context, start, end time.Time) ([]db.LoginCountPerUserInRangeRow, error) {
	return s.repo.LoginCountPerUserInRange(ctx, start, end)
}

func (s *MetricsService) ClientUsageForUser(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]db.ClientUsageForUserInRangeRow, error) {
	return s.repo.ClientUsageForUserInRange(ctx, userID, start, end)
}
