package api

import (
	"encoding/json"
	"time"
)

type PreconditionSupportMode string

const (
	PreconditionSupportNone      PreconditionSupportMode = ""
	PreconditionSupportKeepAlive PreconditionSupportMode = "keepalive"
)

const defaultPreconditionContribution = 1.0

type RepeatingPlan struct {
	Weekdays []int  `json:"weekdays"` // 0-6 (Sunday-Saturday)
	Time     string `json:"time"`     // HH:MM
	Tz       string `json:"tz"`       // timezone in IANA format
	Soc      int    `json:"soc"`      // target soc
	Active   bool   `json:"active"`   // active flag
}

type PlanStrategy struct {
	Continuous               bool                    `json:"continuous"`               // force continuous planning
	Precondition             time.Duration           `json:"precondition"`             // precondition duration in seconds
	PreconditionContribution float64                 `json:"preconditionContribution"` // credited share of the precondition window
	PreconditionSupportMode  PreconditionSupportMode `json:"preconditionSupportMode"`  // runtime support mode during precondition window
}

type planStrategy struct {
	Continuous               bool                    `json:"continuous"`               // force continuous planning
	Precondition             int64                   `json:"precondition"`             // precondition duration in seconds
	PreconditionContribution *float64                `json:"preconditionContribution"` // credited share of the precondition window
	PreconditionSupportMode  PreconditionSupportMode `json:"preconditionSupportMode"`  // runtime support mode during precondition window
}

type persistedPlanStrategy struct {
	Continuous   bool  `json:"continuous"`   // force continuous planning
	Precondition int64 `json:"precondition"` // precondition duration in seconds
}

func DefaultPlanStrategy() PlanStrategy {
	return PlanStrategy{
		PreconditionContribution: defaultPreconditionContribution,
	}
}

func (ps PlanStrategy) PersistedValue() any {
	return persistedPlanStrategy{
		Continuous:   ps.Continuous,
		Precondition: int64(ps.Precondition.Seconds()),
	}
}

func (ps PlanStrategy) MarshalJSON() ([]byte, error) {
	return json.Marshal(planStrategy{
		Continuous:               ps.Continuous,
		Precondition:             int64(ps.Precondition.Seconds()),
		PreconditionContribution: &ps.PreconditionContribution,
		PreconditionSupportMode:  ps.PreconditionSupportMode,
	})
}

func (ps *PlanStrategy) UnmarshalJSON(data []byte) error {
	var res planStrategy
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	contribution := defaultPreconditionContribution
	if res.PreconditionContribution != nil {
		contribution = *res.PreconditionContribution
	}

	*ps = PlanStrategy{
		Continuous:               res.Continuous,
		Precondition:             time.Duration(res.Precondition) * time.Second,
		PreconditionContribution: contribution,
		PreconditionSupportMode:  res.PreconditionSupportMode,
	}

	return nil
}
