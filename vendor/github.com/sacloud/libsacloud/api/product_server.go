package api

import (
	"fmt"

	"github.com/sacloud/libsacloud/sacloud"
)

// ProductServerAPI サーバープランAPI
type ProductServerAPI struct {
	*baseAPI
}

// NewProductServerAPI サーバープランAPI作成
func NewProductServerAPI(client *Client) *ProductServerAPI {
	return &ProductServerAPI{
		&baseAPI{
			client: client,
			// FuncGetResourceURL
			FuncGetResourceURL: func() string {
				return "product/server"
			},
		},
	}
}

// GetBySpec 指定のコア数/メモリサイズ/世代のプランを取得
func (api *ProductServerAPI) GetBySpec(core int, memGB int, gen sacloud.PlanGenerations) (*sacloud.ProductServer, error) {
	plans, err := api.Reset().Find()
	if err != nil {
		return nil, err
	}
	var res sacloud.ProductServer
	var found bool
	for _, plan := range plans.ServerPlans {
		if plan.CPU == core && plan.GetMemoryGB() == memGB {
			if gen == sacloud.PlanDefault || gen == plan.Generation {
				// PlanDefaultの場合は複数ヒットしうる。
				// この場合より新しい世代を優先する。
				if found && plan.Generation <= res.Generation {
					continue
				}
				res = plan
				found = true
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("Server Plan[core:%d, memory:%d, gen:%d] is not found", core, memGB, gen)
	}
	return &res, nil
}

// IsValidPlan 指定のコア数/メモリサイズ/世代のプランが存在し、有効であるか判定
func (api *ProductServerAPI) IsValidPlan(core int, memGB int, gen sacloud.PlanGenerations) (bool, error) {

	productServer, err := api.GetBySpec(core, memGB, gen)

	if err != nil {
		return false, err
	}

	if productServer == nil {
		return false, fmt.Errorf("Server Plan[core:%d, memory:%d, gen:%d] is not found", core, memGB, gen)
	}

	if productServer.Availability != sacloud.EAAvailable {
		return false, fmt.Errorf("Server Plan[core:%d, memory:%d, gen:%d] is not available", core, memGB, gen)
	}

	return true, nil
}
