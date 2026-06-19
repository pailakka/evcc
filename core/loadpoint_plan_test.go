package core

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/keys"
	dbsettings "github.com/evcc-io/evcc/server/db/settings"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type memorySettings struct {
	floats map[string]float64
	times  map[string]time.Time
	json   map[string][]byte
}

func newMemorySettings() *memorySettings {
	return &memorySettings{
		floats: make(map[string]float64),
		times:  make(map[string]time.Time),
		json:   make(map[string][]byte),
	}
}

func (s *memorySettings) SetString(string, string)      {}
func (s *memorySettings) SetInt(string, int64)          {}
func (s *memorySettings) SetFloatPtr(string, *float64)  {}
func (s *memorySettings) SetBool(string, bool)          {}
func (s *memorySettings) String(string) (string, error) { return "", errors.New("not found") }
func (s *memorySettings) Int(string) (int64, error)     { return 0, errors.New("not found") }
func (s *memorySettings) Bool(string) (bool, error)     { return false, errors.New("not found") }

func (s *memorySettings) SetFloat(key string, val float64) {
	s.floats[key] = val
}

func (s *memorySettings) Float(key string) (float64, error) {
	v, ok := s.floats[key]
	if !ok {
		return 0, errors.New("not found")
	}
	return v, nil
}

func (s *memorySettings) SetTime(key string, val time.Time) {
	s.times[key] = val
}

func (s *memorySettings) Time(key string) (time.Time, error) {
	v, ok := s.times[key]
	if !ok {
		return time.Time{}, errors.New("not found")
	}
	return v, nil
}

func (s *memorySettings) SetJson(key string, val any) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	s.json[key] = b
	return nil
}

func (s *memorySettings) Json(key string, res any) error {
	b, ok := s.json[key]
	if !ok {
		return errors.New("not found")
	}
	return json.Unmarshal(b, res)
}

func TestDeparturePowerWindow(t *testing.T) {
	now := time.Date(2026, time.June, 19, 7, 0, 0, 0, time.UTC)
	planTime := now.Add(time.Hour)

	for _, tc := range []struct {
		name     string
		now      time.Time
		planTime time.Time
		duration time.Duration
		active   bool
	}{
		{"zero duration", now, planTime, 0, false},
		{"zero plan time", now, time.Time{}, time.Hour, false},
		{"before start", now.Add(-time.Second), planTime, time.Hour, false},
		{"at start", now, planTime, time.Hour, true},
		{"inside", now.Add(30 * time.Minute), planTime, time.Hour, true},
		{"at plan time", planTime, planTime, time.Hour, false},
		{"after plan time", planTime.Add(time.Second), planTime, time.Hour, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			active, start, end := departurePowerWindow(tc.now, tc.planTime, tc.duration)
			assert.Equal(t, tc.active, active)
			if tc.planTime.IsZero() || tc.duration <= 0 {
				assert.True(t, start.IsZero())
				assert.True(t, end.IsZero())
			} else {
				assert.Equal(t, tc.planTime.Add(-tc.duration), start)
				assert.Equal(t, tc.planTime, end)
			}
		})
	}
}

func TestNextFuturePlanTime(t *testing.T) {
	now := time.Date(2026, time.June, 19, 7, 0, 0, 0, time.UTC)

	assert.Equal(t, now.Add(time.Hour), nextFuturePlanTime(now, []plan{
		{Id: 1, End: now.Add(2 * time.Hour), Soc: 10},
		{Id: 2, End: now.Add(time.Hour), Soc: 80},
		{Id: 3, End: now.Add(-time.Hour), Soc: 90},
	}))
	assert.True(t, nextFuturePlanTime(now, []plan{
		{Id: 1, End: now, Soc: 10},
		{Id: 2, End: now.Add(-time.Hour), Soc: 80},
	}).IsZero())
}

func TestDeparturePowerPlanTimeIncludesReachedSocPlan(t *testing.T) {
	t.Cleanup(config.Reset)
	config.Reset()

	clk := clock.NewMock()
	now := clk.Now()
	planTime := now.Add(time.Hour)

	ctrl := gomock.NewController(t)
	vehicle := api.NewMockVehicle(ctrl)
	vehicle.EXPECT().Capacity().Return(50.0).AnyTimes()
	vehicle.EXPECT().Features().Return(nil).AnyTimes()
	vehicle.EXPECT().OnIdentified().Return(api.ActionConfig{}).AnyTimes()
	var vehicleInstance api.Vehicle = vehicle

	const vehicleName = "departure-power-plan-time"
	assert.NoError(t, config.Vehicles().Add(config.NewStaticDevice(config.Named{Name: vehicleName}, vehicleInstance)))
	dbsettings.SetTime("vehicle."+vehicleName+"."+keys.PlanTime, planTime)
	dbsettings.SetInt("vehicle."+vehicleName+"."+keys.PlanSoc, 70)

	lp := NewLoadpoint(util.NewLogger("foo"), newMemorySettings())
	lp.clock = clk
	lp.vehicle = vehicle
	lp.vehicleSoc = 80

	assert.Equal(t, planTime, lp.departurePowerPlanTime(now))
}

func TestPlannerActiveRetainsReachedEnergyPlanForDeparturePower(t *testing.T) {
	clk := clock.NewMock()
	planTime := clk.Now().Add(time.Hour)

	lp := NewLoadpoint(util.NewLogger("foo"), newMemorySettings())
	lp.clock = clk
	lp.status = api.StatusB
	lp.phases = 1
	lp.planTime = planTime
	lp.planEnergy = 1
	lp.energyMetrics.Update(1)
	lp.planStrategy = api.PlanStrategy{DeparturePower: 30 * time.Minute}

	assert.False(t, lp.plannerActive())
	assert.Equal(t, planTime, lp.planTime)
	assert.Equal(t, 1.0, lp.planEnergy)
}

func TestPlannerActiveClearsReachedEnergyPlanWithoutDeparturePower(t *testing.T) {
	clk := clock.NewMock()
	planTime := clk.Now().Add(time.Hour)

	lp := NewLoadpoint(util.NewLogger("foo"), newMemorySettings())
	lp.clock = clk
	lp.status = api.StatusB
	lp.phases = 1
	lp.planTime = planTime
	lp.planEnergy = 1
	lp.energyMetrics.Update(1)

	assert.False(t, lp.plannerActive())
	assert.True(t, lp.planTime.IsZero())
	assert.Zero(t, lp.planEnergy)
}

func TestDeparturePowerActiveForReachedEnergyPlan(t *testing.T) {
	clk := clock.NewMock()
	planTime := clk.Now().Add(time.Hour)

	lp := NewLoadpoint(util.NewLogger("foo"), newMemorySettings())
	lp.clock = clk
	lp.status = api.StatusB
	lp.phases = 1
	lp.planTime = planTime
	lp.planEnergy = 1
	lp.energyMetrics.Update(1)
	lp.planStrategy = api.PlanStrategy{DeparturePower: 30 * time.Minute}

	assert.False(t, lp.departurePowerActive())

	clk.Add(30 * time.Minute)
	assert.True(t, lp.departurePowerActive())

	clk.Add(30 * time.Minute)
	assert.False(t, lp.plannerActive())
	assert.True(t, lp.planTime.IsZero())
	assert.True(t, lp.departurePowerActive())

	lp.status = api.StatusA
	assert.False(t, lp.departurePowerActive())
}

func TestDeparturePowerInactiveWhenDisconnected(t *testing.T) {
	clk := clock.NewMock()

	lp := NewLoadpoint(util.NewLogger("foo"), newMemorySettings())
	lp.clock = clk
	lp.status = api.StatusA
	lp.planTime = clk.Now().Add(time.Hour)
	lp.planEnergy = 1
	lp.planStrategy = api.PlanStrategy{DeparturePower: time.Hour}

	assert.False(t, lp.departurePowerActive())
}

func TestKeepPlanForDeparturePower(t *testing.T) {
	clk := clock.NewMock()
	lp := NewLoadpoint(util.NewLogger("foo"), newMemorySettings())
	lp.clock = clk
	lp.status = api.StatusB

	strategy := api.PlanStrategy{DeparturePower: time.Hour}
	assert.True(t, lp.keepPlanForDeparturePower(clk.Now().Add(time.Minute), strategy))
	assert.False(t, lp.keepPlanForDeparturePower(clk.Now(), strategy))
	assert.False(t, lp.keepPlanForDeparturePower(clk.Now().Add(time.Minute), api.PlanStrategy{}))

	lp.status = api.StatusA
	assert.False(t, lp.keepPlanForDeparturePower(clk.Now().Add(time.Minute), strategy))
}

func TestReachedEnergyPlanClearsStoredSettingsWithoutDeparturePower(t *testing.T) {
	clk := clock.NewMock()
	settings := newMemorySettings()
	planTime := clk.Now().Add(time.Hour)

	lp := NewLoadpoint(util.NewLogger("foo"), settings)
	lp.clock = clk
	lp.status = api.StatusB
	lp.phases = 1
	lp.planTime = planTime
	lp.planEnergy = 1
	lp.energyMetrics.Update(1)

	assert.False(t, lp.plannerActive())

	storedPlanTime, err := settings.Time(keys.PlanTime)
	assert.NoError(t, err)
	assert.True(t, storedPlanTime.IsZero())

	storedEnergy, err := settings.Float(keys.PlanEnergy)
	assert.NoError(t, err)
	assert.Zero(t, storedEnergy)
}
