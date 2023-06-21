package kubeapi

import (
	"kube-rise/pkg/mocks"
	"kube-rise/pkg/server/infrastructure"
)

type MetaKubeAPIServer struct {
	GetApi                   infrastructure.Endpoint
	GetApis                  infrastructure.Endpoint
	GetClusterAPIDescription infrastructure.Endpoint
	GetV1Api                 infrastructure.Endpoint
	GetAutoscalingApi        infrastructure.Endpoint
}

func NewMetaKubeAPIServer() *MetaKubeAPIServer {
	var server = &MetaKubeAPIServer{}
	server.GetApi = infrastructure.HandleRequest(mocks.GetApiVersions)
	server.GetApis = infrastructure.HandleRequest(mocks.GetApis)
	server.GetClusterAPIDescription = infrastructure.HandleRequest(mocks.GetClusterAPIDescription)
	server.GetV1Api = infrastructure.HandleRequest(mocks.GetV1Api)
	server.GetAutoscalingApi = infrastructure.HandleRequest(mocks.GetAutoscalingApi)
	return server
}
