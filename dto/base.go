package dto

import (
	"gorm.io/gorm"
)

const MaxPageSize = 200

type Pagination struct {
	Is       bool
	Page     int `form:"page"`
	PageSize int `form:"page_size" binding:"required_with=Page"`
	Offset   int
	Limit    int
}

// PagedDB TODO: refactor, merge Page, Offset usage
func (p *Pagination) PagedDB(tx *gorm.DB) *gorm.DB {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = MaxPageSize
	}
	if p.PageSize == 0 {
		p.PageSize = MaxPageSize
	}

	if p.Offset == 0 {
		p.Offset = (p.Page - 1) * p.PageSize
		p.Limit = p.PageSize
	}

	return tx.Offset(p.Offset).Limit(p.Limit)
}

type BaseRsp struct {
	Message string `json:"message"`
}

type PagedRsp struct {
	Count    int64 `json:"count"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func (dto *PagedRsp) FillPagination(pagination Pagination) {
	dto.Page = pagination.Page
	dto.PageSize = pagination.PageSize
}
