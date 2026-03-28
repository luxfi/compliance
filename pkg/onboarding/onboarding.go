// Package onboarding provides the 5-step investor onboarding application flow.
package onboarding

import (
	"time"

	"github.com/luxfi/compliance/pkg/types"
)

// NewApplicationSteps returns the 5 default onboarding steps.
func NewApplicationSteps() []types.ApplicationStep {
	return []types.ApplicationStep{
		{Step: 1, Name: "Basic Info & Contact", Status: "pending"},
		{Step: 2, Name: "Identity Verification", Status: "pending"},
		{Step: 3, Name: "Document Upload", Status: "pending"},
		{Step: 4, Name: "Compliance Screening", Status: "pending"},
		{Step: 5, Name: "Review & Submit", Status: "pending"},
	}
}

// IsTerminalStatus returns true if the application is in a final state
// where no further step modifications are allowed.
func IsTerminalStatus(status types.ApplicationStatus) bool {
	return status == types.AppApproved || status == types.AppRejected || status == types.AppSubmitted
}

// MarkStepCompleted sets a step to completed.
func MarkStepCompleted(app *types.Application, step int) {
	for i := range app.Steps {
		if app.Steps[i].Step == step {
			app.Steps[i].Status = "completed"
			app.Steps[i].CompletedAt = time.Now().UTC()
			return
		}
	}
}

// MarkStepFailed sets a step to failed.
func MarkStepFailed(app *types.Application, step int) {
	for i := range app.Steps {
		if app.Steps[i].Step == step {
			app.Steps[i].Status = "failed"
			return
		}
	}
}
