// Package tickler fires due reminders as in-app notifications on an interval.
package tickler

import (
	"context"
	"log"
	"time"

	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
)

// Runner periodically converts due reminders into system notifications.
type Runner struct {
	Interval time.Duration
}

// Start blocks until ctx is cancelled, firing due reminders on each interval.
func (r Runner) Start(ctx context.Context) {
	if r.Interval <= 0 {
		r.Interval = 15 * time.Minute
	}
	t := time.NewTicker(r.Interval)
	defer t.Stop()
	log.Printf("tickler: started (interval=%s)", r.Interval)
	for {
		select {
		case <-ctx.Done():
			log.Print("tickler: stopped")
			return
		case <-t.C:
			r.runOnce(ctx)
		}
	}
}

// runOnce fires all currently-due reminders.
func (r Runner) runOnce(ctx context.Context) {
	rows, err := model.Queries.ListDueReminders(ctx)
	if err != nil {
		log.Printf("tickler: ListDueReminders: %v", err)
		return
	}
	for _, rem := range rows {
		if err := r.fire(ctx, rem); err != nil {
			log.Printf("tickler: fire reminder %d: %v", rem.ID, err)
		}
	}
}

// fire creates an in-app notification for the reminder and marks it completed.
func (r Runner) fire(ctx context.Context, rem dbgen.Reminder) error {
	patientID := int64(0)
	if rem.Patient.Valid {
		patientID = rem.Patient.Int64
	}
	if _, err := model.Queries.CreateNotification(ctx, dbgen.CreateNotificationParams{
		UserID:    rem.User,
		Text:      rem.Title,
		Action:    "reminder",
		Module:    "reminders",
		PatientID: patientID,
	}); err != nil {
		return err
	}
	return model.Queries.UpdateReminderStatus(ctx, dbgen.UpdateReminderStatusParams{
		ReminderID: rem.ID,
		Status:     "completed",
	})
}
