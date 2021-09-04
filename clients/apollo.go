package clients

import (
	"fmt"
	"net/http"
)

type apollo struct {
	baseRest
	Token string
}

func Apollo() *apollo {
	if RestConfigs.ApolloConfig == nil {
		panic("no apollo config")
	}

	apolloConfig := defaultHTTPConfig
	apolloConfig.Headers["Authorization"] = RestConfigs.ApolloConfig.Token

	return &apollo{
		baseRest: baseRest{
			Scheme:  RestConfigs.ApolloConfig.Scheme,
			Host:    RestConfigs.ApolloConfig.Hostname,
			URIBase: "openapi/v1",
			Config:  apolloConfig,
		},
	}
}

type GetNamespaceInfoReq struct {
	Env           string
	AppID         string
	ClusterName   string
	NamespaceName string
}

func (a *apollo) GetNamespaceInfo(req GetNamespaceInfoReq) (rsp map[string]interface{}, err error) {
	subURL := fmt.Sprintf("envs/%s/apps/%s/clusters/%s/namespaces/%s/",
		req.Env, req.AppID, req.ClusterName, req.NamespaceName)
	err = a.Json(http.MethodPost, subURL, &req, &rsp)
	return
}
