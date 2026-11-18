// Package report_scheduler supplies dashboard HTTP adapters for the independently
// versioned scheduler module. Business logic and persistence live in that module.
package report_scheduler

import (
	"database/sql"
	"time"

	scheduler "github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler"
)

type Address = scheduler.Address
type Attachment = scheduler.Attachment
type Message = scheduler.Message
type User = scheduler.User
type Notification = scheduler.Notification
type EmailDelivery = scheduler.EmailDelivery
type UserLookup = scheduler.UserLookup
type ArtifactStorage = scheduler.ArtifactStorage
type Audit = scheduler.Audit
type Notifications = scheduler.Notifications
type ModuleResponse = scheduler.ModuleResponse
type HealthBIReport = scheduler.HealthBIReport
type HealthBIParameter = scheduler.HealthBIParameter
type HealthContext = scheduler.HealthContext
type ScheduleRecipient = scheduler.ScheduleRecipient
type RecipientPreview = scheduler.RecipientPreview
type RecipientOption = scheduler.RecipientOption
type OutputConfig = scheduler.OutputConfig
type ScheduleTiming = scheduler.ScheduleTiming
type ResolvedPeriod = scheduler.ResolvedPeriod
type GenerateReportRequest = scheduler.GenerateReportRequest
type HealthBIJob = scheduler.HealthBIJob
type Schedule = scheduler.Schedule
type CreateScheduleRequest = scheduler.CreateScheduleRequest
type UpdateScheduleRequest = scheduler.UpdateScheduleRequest
type ListOptions = scheduler.ListOptions
type Execution = scheduler.Execution
type Delivery = scheduler.Delivery
type ExecutionDetail = scheduler.ExecutionDetail
type Artifact = scheduler.Artifact
type SchedulerStatusCount = scheduler.SchedulerStatusCount
type SchedulerOverview = scheduler.SchedulerOverview
type PortalReport = scheduler.PortalReport
type Service = scheduler.Service
type Repository = scheduler.Repository
type WorkerService = scheduler.WorkerService
type Worker = scheduler.Worker
type HealthBIClient = scheduler.HealthBIClient
type HealthBIRequestError = scheduler.HealthBIRequestError
type HealthHierarchy = scheduler.HealthHierarchy

var ErrScheduleNotFound = scheduler.ErrScheduleNotFound

func NewService(repo Repository, provider HealthBIClient, email EmailDelivery, users UserLookup, storage ArtifactStorage, hierarchy HealthHierarchy) *Service {
	return scheduler.NewService(repo, provider, email, users, storage, hierarchy)
}
func NewRepository(db *sql.DB) Repository { return scheduler.NewRepository(db) }
func NewWorker(service WorkerService, interval time.Duration, log func(...any)) *Worker {
	return scheduler.NewWorker(service, interval, log)
}
func NewKeycloakHealthBIClient(baseURL, tokenURL, clientID, clientSecret string, timeout time.Duration) HealthBIClient {
	return scheduler.NewKeycloakHealthBIClient(baseURL, tokenURL, clientID, clientSecret, timeout)
}
func NewPostgresHealthHierarchy(db *sql.DB) HealthHierarchy {
	return scheduler.NewPostgresHealthHierarchy(db)
}
func CalculateNextRun(frequency, timezone string, timing ScheduleTiming, after time.Time) (time.Time, error) {
	return scheduler.CalculateNextRun(frequency, timezone, timing, after)
}
func ResolveReportingPeriod(strategy, timezone string, reference time.Time) (ResolvedPeriod, error) {
	return scheduler.ResolveReportingPeriod(strategy, timezone, reference)
}
