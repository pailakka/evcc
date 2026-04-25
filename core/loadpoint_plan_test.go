package core

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/vehicle"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type planTestVehicle struct {
	capacity float64
}

func (v *planTestVehicle) Soc() (float64, error) {
	return 0, nil
}

func (v *planTestVehicle) Capacity() float64 {
	return v.capacity
}

func (v *planTestVehicle) Icon() string {
	return ""
}

func (v *planTestVehicle) Features() []api.Feature {
	return nil
}

func (v *planTestVehicle) Phases() int {
	return 1
}

func (v *planTestVehicle) GetTitle() string {
	return "test vehicle"
}

func (v *planTestVehicle) SetTitle(string) {
}

func (v *planTestVehicle) Identifiers() []string {
	return nil
}

func (v *planTestVehicle) OnIdentified() api.ActionConfig {
	return api.ActionConfig{}
}

func newPlanTestLoadpoint(t *testing.T, now time.Time) (*Loadpoint, *clock.Mock, *planTestVehicle) {
	t.Helper()

	config.Reset()
	t.Cleanup(config.Reset)
	Voltage = 230

	clk := clock.NewMock()
	clk.Set(now)

	lp := NewLoadpoint(util.NewLogger("foo"), nil)
	lp.clock = clk
	lp.status = api.StatusB
	lp.vehicleSoc = 80

	v := &planTestVehicle{capacity: 60}
	lp.vehicle = v

	err := config.Vehicles().Add(config.NewStaticDevice[api.Vehicle](config.Named{Name: "test-vehicle"}, v))
	require.NoError(t, err)

	return lp, clk, v
}

func TestPlannerActiveKeepAliveInsidePreconditionWindow(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	lp, _, v := newPlanTestLoadpoint(t, now)

	strategy := api.DefaultPlanStrategy()
	strategy.Precondition = 30 * time.Minute
	strategy.PreconditionSupportMode = api.PreconditionSupportKeepAlive
	require.NoError(t, vehicle.Settings(lp.log, v).SetPlanStrategy(strategy))

	planTime := now.Add(20 * time.Minute)
	lp.lockPlanGoal(planTime, 80, 1)
	lp.hardLimitSoc = 80

	active := lp.plannerActive()

	assert.True(t, active)
	assert.True(t, lp.planActive)
}

func TestPlannerActiveKeepAliveRequiresHardLimit(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	lp, _, v := newPlanTestLoadpoint(t, now)

	strategy := api.DefaultPlanStrategy()
	strategy.Precondition = 30 * time.Minute
	strategy.PreconditionSupportMode = api.PreconditionSupportKeepAlive
	require.NoError(t, vehicle.Settings(lp.log, v).SetPlanStrategy(strategy))

	lp.lockPlanGoal(now.Add(20*time.Minute), 80, 1)

	active := lp.plannerActive()

	assert.False(t, active)
	assert.False(t, lp.planActive)
}

func TestPlannerActiveKeepAliveEndsAtPlanTime(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	lp, clk, v := newPlanTestLoadpoint(t, now)

	strategy := api.DefaultPlanStrategy()
	strategy.Precondition = 30 * time.Minute
	strategy.PreconditionSupportMode = api.PreconditionSupportKeepAlive
	require.NoError(t, vehicle.Settings(lp.log, v).SetPlanStrategy(strategy))

	planTime := now.Add(20 * time.Minute)
	lp.lockPlanGoal(planTime, 80, 1)
	lp.hardLimitSoc = 80

	assert.True(t, lp.plannerActive())

	clk.Set(planTime)

	assert.False(t, lp.plannerActive())
	assert.False(t, lp.planActive)
}

func TestPlannerActiveKeepAliveOutsidePreconditionWindow(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	lp, _, v := newPlanTestLoadpoint(t, now)

	strategy := api.DefaultPlanStrategy()
	strategy.Precondition = 30 * time.Minute
	strategy.PreconditionSupportMode = api.PreconditionSupportKeepAlive
	require.NoError(t, vehicle.Settings(lp.log, v).SetPlanStrategy(strategy))

	lp.lockPlanGoal(now.Add(45*time.Minute), 80, 1)
	lp.hardLimitSoc = 80

	active := lp.plannerActive()

	assert.False(t, active)
	assert.False(t, lp.planActive)
}

func TestPlannerActiveGoal100Unchanged(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	lp, _, v := newPlanTestLoadpoint(t, now)

	require.NoError(t, vehicle.Settings(lp.log, v).SetPlanStrategy(api.DefaultPlanStrategy()))

	lp.vehicleSoc = 100
	lp.lockPlanGoal(now.Add(2*time.Hour), 100, 1)
	lp.planActive = true

	active := lp.plannerActive()

	assert.True(t, active)
	assert.True(t, lp.planActive)
}
