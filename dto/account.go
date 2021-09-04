package dto

import (
	"github.com/jinzhu/copier"

	"go-cygnus/clients"
	"go-cygnus/models"
	"go-cygnus/utils/db"
)

type ListAccountReq struct{}

type ListAccountRsp struct {
	PagedRsp
	Result []models.Account `json:"result"`
}

func (dto *ListAccountReq) List(pagination Pagination) (rsp ListAccountRsp, err error) {
	rsp.FillPagination(pagination)

	err = db.Engine.Model(&models.Account{}).Count(&rsp.Count).Offset(
		pagination.Offset).Limit(pagination.Limit).Find(&rsp.Result).Error
	return
}

type AddAccountReq struct {
	AppID         string `json:"app_id"`
	Env           string `json:"env"`
	ClusterName   string `json:"cluster_name"`
	NamespaceName string `json:"namespace_name"`
}

type AddAccountRsp struct{}

func (dto *AddAccountReq) Add() (rsp AddAccountRsp, err error) {
	var apolloReq clients.GetNamespaceInfoReq

	if err = copier.Copy(&apolloReq, dto); err != nil {
		return
	}

	if _, err = clients.Apollo().GetNamespaceInfo(apolloReq); err != nil {
		return
	}

	return
}
