package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanStrategyUnmarshalDefaults(t *testing.T) {
	var strategy PlanStrategy

	err := json.Unmarshal([]byte(`{"continuous":true,"precondition":1800}`), &strategy)
	require.NoError(t, err)

	assert.Equal(t, 30*time.Minute, strategy.Precondition)
	assert.True(t, strategy.Continuous)
	assert.Equal(t, 1.0, strategy.PreconditionContribution)
	assert.Equal(t, PreconditionSupportNone, strategy.PreconditionSupportMode)
}

func TestPlanStrategyRoundTrip(t *testing.T) {
	strategy := PlanStrategy{
		Continuous:               true,
		Precondition:             45 * time.Minute,
		PreconditionContribution: 0,
		PreconditionSupportMode:  PreconditionSupportKeepAlive,
	}

	data, err := json.Marshal(strategy)
	require.NoError(t, err)

	var decoded PlanStrategy
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, strategy, decoded)
}

func TestPlanStrategyPersistedValueStaysLegacyCompatible(t *testing.T) {
	strategy := PlanStrategy{
		Continuous:               true,
		Precondition:             45 * time.Minute,
		PreconditionContribution: 0.25,
		PreconditionSupportMode:  PreconditionSupportKeepAlive,
	}

	data, err := json.Marshal(strategy.PersistedValue())
	require.NoError(t, err)

	assert.JSONEq(t, `{"continuous":true,"precondition":2700}`, string(data))
}
