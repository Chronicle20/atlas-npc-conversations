package configuration

import (
	"atlas-npc-conversations/configuration/tenant"
	"atlas-npc-conversations/rest"
	"github.com/Chronicle20/atlas-rest/requests"
)

const (
	Resource   = "configurations"
	AllTenants = Resource + "/tenants"
)

func getBaseRequest() string {
	return requests.RootUrl("CONFIGURATIONS")
}

func requestAllTenants() requests.Request[[]tenant.RestModel] {
	return rest.MakeGetRequest[[]tenant.RestModel](getBaseRequest() + AllTenants)
}
