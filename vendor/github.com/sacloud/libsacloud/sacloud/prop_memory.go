package sacloud

// propMemoryMB サイズ(MB)内包型
type propMemoryMB struct {
	MemoryMB int `json:",omitempty"` // サイズ(MB単位)
}

// GetMemoryMB サイズ(MB単位) 取得
func (p *propMemoryMB) GetMemoryMB() int {
	return p.MemoryMB
}

// GetMemoryGB サイズ(GB単位) 取得
func (p *propMemoryMB) GetMemoryGB() int {
	if p.MemoryMB <= 0 {
		return 0
	}
	return p.MemoryMB / 1024
}

// SetMemoryGB サイズ(GB単位) 設定
func (p *propMemoryMB) SetMemoryGB(memoryGB int) {
	p.MemoryMB = memoryGB * 1024
}
