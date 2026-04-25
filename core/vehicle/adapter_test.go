package vehicle

import (
	"testing"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/keys"
	dbsettings "github.com/evcc-io/evcc/server/db/settings"
	"github.com/evcc-io/evcc/util"
	"github.com/stretchr/testify/assert"
)

func TestPlanStrategySettingsMigration(t *testing.T) {
	v := &adapter{
		log:  util.NewLogger("foo"),
		name: "migration-test",
	}

	publishCalls := 0
	prevPublish := Publish
	Publish = func() {
		publishCalls++
	}
	t.Cleanup(func() {
		Publish = prevPublish
	})

	keyPrefix := v.key()
	dbsettings.SetString(keyPrefix+keys.PlanStrategy, `{"continuous":true,"precondition":900,"preconditionContribution":0.25,"preconditionSupportMode":"keepalive"}`)

	assert.Equal(t, api.PlanStrategy{
		Continuous:               true,
		Precondition:             15 * time.Minute,
		PreconditionContribution: 0.25,
		PreconditionSupportMode:  api.PreconditionSupportKeepAlive,
	}, v.GetPlanStrategy())

	rawStrategy, err := dbsettings.String(keyPrefix + keys.PlanStrategy)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"continuous":true,"precondition":900}`, rawStrategy)

	contribution, err := dbsettings.Float(keyPrefix + keys.PlanContribution)
	assert.NoError(t, err)
	assert.Equal(t, 0.25, contribution)

	supportMode, err := dbsettings.String(keyPrefix + keys.PlanSupportMode)
	assert.NoError(t, err)
	assert.Equal(t, "keepalive", supportMode)
	assert.Zero(t, publishCalls)
}
