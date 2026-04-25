package charger

import (
	"testing"

	"github.com/evcc-io/evcc/api"
	"github.com/stretchr/testify/assert"
)

func TestVehicleControlFeature(t *testing.T) {
	assert.Contains(t, (&VehicleApi{}).Features(), api.VehicleControl)
	assert.Contains(t, (&Twc3{}).Features(), api.VehicleControl)
}
