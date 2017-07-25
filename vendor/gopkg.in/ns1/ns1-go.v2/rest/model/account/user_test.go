package account

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalUsers(t *testing.T) {
	d := []byte(`[
  {
    "permissions": {},
    "teams": [],
    "email": "support@nsone.net",
    "last_access": 1376325771.0,
    "notify": {
      "billing": true
    },
    "name": "API Example",
    "username": "apiexample"
  },
  {
    "permissions": {
      "dns": {
        "view_zones": true,
        "manage_zones": true,
        "zones_allow_by_default": false,
        "zones_deny": [],
        "zones_allow": ["example.com"]
      },
      "data": {
        "push_to_datafeeds": false,
        "manage_datasources": false,
        "manage_datafeeds": false
      },
      "account": {
        "manage_payment_methods": false,
        "manage_plan": false,
        "manage_teams": false,
        "manage_apikeys": false,
        "manage_account_settings": false,
        "view_activity_log": false,
        "view_invoices": false,
        "manage_users": false
      },
      "monitoring": {
        "manage_lists": false,
        "manage_jobs": false,
        "view_jobs": false
      }
    },
    "teams": ["520422919f782d37dffb588a"],
    "email": "newuser@example.com",
    "last_access": null,
    "notify": {
      "billing": true
    },
    "name": "New User",
    "username": "newuser"
  }
]
`)
	ul := []*User{}
	if err := json.Unmarshal(d, &ul); err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(ul), 2, "Userlist should have 2 users")

	u := ul[0]
	assert.Equal(t, u.TeamIDs, []string{}, "User should have empty teams")
	assert.Equal(t, u.Email, "support@nsone.net", "User wrong email")
	assert.Equal(t, u.LastAccess, 1376325771.0, "User wrong last access")
	assert.Equal(t, u.Name, "API Example", "User wrong name")
	assert.Equal(t, u.Username, "apiexample", "User wrong username")
	assert.Equal(t, u.Notify, NotificationSettings{true}, "User wrong notify")
	assert.Equal(t, u.Permissions, PermissionsMap{}, "User should have empty permissions")

	u2 := ul[1]
	assert.Equal(t, u2.TeamIDs, []string{"520422919f782d37dffb588a"}, "User should have empty teams")
	assert.Equal(t, u2.Email, "newuser@example.com", "User wrong email")
	assert.Equal(t, u2.LastAccess, 0.0, "User wrong last access")
	assert.Equal(t, u2.Name, "New User", "User wrong name")
	assert.Equal(t, u2.Username, "newuser", "User wrong username")
	assert.Equal(t, u.Notify, NotificationSettings{true}, "User wrong notify")

	permMap := PermissionsMap{
		DNS: PermissionsDNS{
			ViewZones:           true,
			ManageZones:         true,
			ZonesAllowByDefault: false,
			ZonesDeny:           []string{},
			ZonesAllow:          []string{"example.com"},
		},
		Data: PermissionsData{
			PushToDatafeeds:   false,
			ManageDatasources: false,
			ManageDatafeeds:   false,
		},
		Account: PermissionsAccount{
			ManagePaymentMethods:  false,
			ManagePlan:            false,
			ManageTeams:           false,
			ManageApikeys:         false,
			ManageAccountSettings: false,
			ViewActivityLog:       false,
			ViewInvoices:          false,
			ManageUsers:           false,
		},
		Monitoring: PermissionsMonitoring{
			ManageLists: false,
			ManageJobs:  false,
			ViewJobs:    false,
		},
	}
	assert.Equal(t, u2.Permissions, permMap, "User wrong permissions")
}
