package metrics

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
)

type metricsRepository struct {
	db     db.Store
	config *config.Config
	logger *logger.Logger
}

func NewMetricsRepository(
	cfg *config.Config,
	store db.Store,
	log logger.Logger,
) MetricsRepository {

	return &metricsRepository{
		db:     store,
		config: cfg,
		logger: &log,
	}
}

func (r *metricsRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.db.CountUsers(ctx)
}

func (r *metricsRepository) CountDisabledUsers(ctx context.Context) (int64, error) {
	return r.db.CountDisabledUsers(ctx)
}

func (r *metricsRepository) CountActiveUsers(ctx context.Context) (int64, error) {
	return r.db.CountActiveUsers(ctx)
}

func (r *metricsRepository) CountActiveUsersInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return r.db.CountActiveUsersInRange(ctx, db.CountActiveUsersInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) ActiveUsersToday(ctx context.Context) (int64, error) {
	return r.db.ActiveUsersToday(ctx)
}

func (r *metricsRepository) ActiveUsersThisWeek(ctx context.Context) (int64, error) {
	return r.db.ActiveUsersThisWeek(ctx)
}

func (r *metricsRepository) LoginTrend(ctx context.Context) ([]db.LoginTrendRow, error) {
	return r.db.LoginTrend(ctx)
}

func (r *metricsRepository) LoginTrendByDay(ctx context.Context, start, end time.Time) ([]db.LoginTrendByDayRow, error) {
	return r.db.LoginTrendByDay(ctx, db.LoginTrendByDayParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) LoginSuccessFailureInRange(ctx context.Context, start, end time.Time) (db.LoginSuccessFailureInRangeRow, error) {
	return r.db.LoginSuccessFailureInRange(ctx, db.LoginSuccessFailureInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) TotalLoginsInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return r.db.TotalLoginsInRange(ctx, db.TotalLoginsInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) ApproximateActiveSessions(ctx context.Context) (int64, error) {
	return r.db.ApproximateActiveSessions(ctx)
}

func (r *metricsRepository) CountFailedLogins(ctx context.Context) (int64, error) {
	return r.db.CountFailedLogins(ctx)
}

func (r *metricsRepository) CountFailedLoginsInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return r.db.CountFailedLoginsInRange(ctx, db.CountFailedLoginsInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) CountPasswordResetsInRange(ctx context.Context, start, end time.Time) (int64, error) {
	return r.db.CountPasswordResetsInRange(ctx, db.CountPasswordResetsInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) SuspiciousLoginsInRange(ctx context.Context, start, end time.Time, homeCountry string) ([]db.SuspiciousLoginsInRangeRow, error) {
	return r.db.SuspiciousLoginsInRange(ctx, db.SuspiciousLoginsInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
		HomeCountry: homeCountry,
	})
}

func (r *metricsRepository) CountClients(ctx context.Context) (int64, error) {
	return r.db.CountClients(ctx)
}

func (r *metricsRepository) CountEnabledClients(ctx context.Context) (int64, error) {
	return r.db.CountEnabledClients(ctx)
}

func (r *metricsRepository) MostActiveClients(ctx context.Context) ([]db.MostActiveClientsRow, error) {
	return r.db.MostActiveClients(ctx)
}

func (r *metricsRepository) MostAccessedClients(ctx context.Context, start, end time.Time, rowLimit int32) ([]db.MostAccessedClientsRow, error) {
	return r.db.MostAccessedClients(ctx, db.MostAccessedClientsParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
		RowLimit: rowLimit,
	})
}

func (r *metricsRepository) LoginCountForClientInRange(ctx context.Context, clientID string, start, end time.Time) (int64, error) {
	return r.db.LoginCountForClientInRange(ctx, db.LoginCountForClientInRangeParams{
		ClientID: clientID,
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) ActiveUsersPerClientToday(ctx context.Context) ([]db.ActiveUsersPerClientTodayRow, error) {
	return r.db.ActiveUsersPerClientToday(ctx)
}

func (r *metricsRepository) NewClientsInRange(
	ctx context.Context,
	start, end time.Time,
) ([]db.NewClientsInRangeRow, error) {
	return r.db.NewClientsInRange(ctx, db.NewClientsInRangeParams{
		StartTime: start,
		EndTime:   end,
	})
}

func (r *metricsRepository) RecentlyCreatedClients(
	ctx context.Context,
	rowLimit int32,
) ([]db.RecentlyCreatedClientsRow, error) {
	return r.db.RecentlyCreatedClients(ctx, rowLimit)
}

func (r *metricsRepository) NewUsersInRange(ctx context.Context, start, end time.Time) ([]db.User, error) {
	return r.db.NewUsersInRange(ctx, db.NewUsersInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) NewUsersTrend(ctx context.Context, start, end time.Time) ([]db.NewUsersTrendRow, error) {
	return r.db.NewUsersTrend(ctx, db.NewUsersTrendParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) CountNewUsersToday(ctx context.Context) (int64, error) {
	return r.db.CountNewUsersToday(ctx)
}

func (r *metricsRepository) CountNewUsersThisWeek(ctx context.Context) (int64, error) {
	return r.db.CountNewUsersThisWeek(ctx)
}

func (r *metricsRepository) NeverLoggedInUsers(ctx context.Context) ([]db.User, error) {
	return r.db.NeverLoggedInUsers(ctx)
}

func (r *metricsRepository) LastLoginForUser(ctx context.Context, userID uuid.UUID) (sql.NullTime, error) {
	return r.db.LastLoginForUser(ctx, userID)
}

func (r *metricsRepository) LastLoginForAllUsers(ctx context.Context) ([]db.LastLoginForAllUsersRow, error) {
	return r.db.LastLoginForAllUsers(ctx)
}

func (r *metricsRepository) LoginCountPerUserInRange(ctx context.Context, start, end time.Time) ([]db.LoginCountPerUserInRangeRow, error) {
	return r.db.LoginCountPerUserInRange(ctx, db.LoginCountPerUserInRangeParams{
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}

func (r *metricsRepository) ClientUsageForUserInRange(
	ctx context.Context,
	userID uuid.UUID,
	start, end time.Time,
) ([]db.ClientUsageForUserInRangeRow, error) {

	return r.db.ClientUsageForUserInRange(ctx, db.ClientUsageForUserInRangeParams{
		UserID: userID,
		StartTime: sql.NullTime{
			Time:  start,
			Valid: true,
		},
		EndTime: sql.NullTime{
			Time:  end,
			Valid: true,
		},
	})
}
