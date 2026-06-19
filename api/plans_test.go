package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanStrategyJson(t *testing.T) {
	strategy := PlanStrategy{
		Continuous:     true,
		Precondition:   30 * time.Minute,
		DeparturePower: time.Hour,
	}

	b, err := json.Marshal(strategy)
	require.NoError(t, err)
	assert.JSONEq(t, `{"continuous":true,"precondition":1800,"departurePower":3600}`, string(b))

	var decoded PlanStrategy
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, strategy, decoded)
}

func TestPlanStrategyJsonBackwardsCompatibility(t *testing.T) {
	var strategy PlanStrategy
	require.NoError(t, json.Unmarshal([]byte(`{"continuous":true,"precondition":900}`), &strategy))

	assert.True(t, strategy.Continuous)
	assert.Equal(t, 15*time.Minute, strategy.Precondition)
	assert.Zero(t, strategy.DeparturePower)
}
