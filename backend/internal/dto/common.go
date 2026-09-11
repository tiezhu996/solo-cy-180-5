// Package dto 定义接口出入参结构。
package dto

// PageParams 统一分页参数。
type PageParams struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// Normalize 归一化分页参数。
func (p *PageParams) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 20
	}
}
