package sacloud

// propServerPlan サーバープラン内包型
type propServerPlan struct {
	ServerPlan *ProductServer `json:",omitempty"` // サーバープラン
}

// GetServerPlan サーバープラン取得
func (p *propServerPlan) GetServerPlan() *ProductServer {
	return p.ServerPlan
}

// SetServerPlan サーバープラン設定
func (p *propServerPlan) SetServerPlan(plan *ProductServer) {
	p.ServerPlan = plan
}

// SetServerPlanByID サーバープラン設定
func (p *propServerPlan) SetServerPlanByID(planID string) {
	if p.ServerPlan == nil {
		p.ServerPlan = &ProductServer{}
	}
	p.ServerPlan.Resource = NewResourceByStringID(planID)
}

// SetServerPlanByValue サーバープラン設定(値指定)
func (p *propServerPlan) SetServerPlanByValue(cpu int, memoryGB int, gen PlanGenerations) {
	plan := &ProductServer{}
	plan.CPU = cpu
	plan.SetMemoryGB(memoryGB)
	plan.Generation = gen
	p.ServerPlan = plan
}

// GetCPU CPUコア数 取得
func (p *propServerPlan) GetCPU() int {
	if p.ServerPlan == nil {
		return -1
	}

	return p.ServerPlan.GetCPU()
}

// GetMemoryMB メモリ(MB) 取得
func (p *propServerPlan) GetMemoryMB() int {
	if p.ServerPlan == nil {
		return -1
	}

	return p.ServerPlan.GetMemoryMB()
}

// GetMemoryGB メモリ(GB) 取得
func (p *propServerPlan) GetMemoryGB() int {
	if p.ServerPlan == nil {
		return -1
	}

	return p.ServerPlan.GetMemoryGB()
}

func (p *propServerPlan) SetMemoryGB(memoryGB int) {
	p.ServerPlan.SetMemoryGB(memoryGB)
}
