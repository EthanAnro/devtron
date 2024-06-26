package client

import (
	"github.com/gorilla/mux"
)

type HelmAppRouter interface {
	InitAppListRouter(helmRouter *mux.Router)
}
type HelmAppRouterImpl struct {
	helmAppRestHandler HelmAppRestHandler
}

func NewHelmAppRouterImpl(helmAppRestHandler HelmAppRestHandler) *HelmAppRouterImpl {
	return &HelmAppRouterImpl{
		helmAppRestHandler: helmAppRestHandler,
	}
}

func (impl *HelmAppRouterImpl) InitAppListRouter(helmRouter *mux.Router) {
	helmRouter.Path("").Queries("clusterIds", "{clusterIds}").
		HandlerFunc(impl.helmAppRestHandler.ListApplications).Methods("GET")
	helmRouter.Path("/app").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.GetApplicationDetail).Methods("GET")

	helmRouter.Path("/app/save-telemetry").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.SaveHelmAppDetailsViewedTelemetryData).Methods("GET")

	helmRouter.Path("/hibernate").HandlerFunc(impl.helmAppRestHandler.Hibernate).Methods("POST")
	helmRouter.Path("/unhibernate").HandlerFunc(impl.helmAppRestHandler.UnHibernate).Methods("POST")

	// GetReleaseInfo used only for external apps
	helmRouter.Path("/release-info").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.GetReleaseInfo).Methods("GET")

	helmRouter.Path("/desired-manifest").HandlerFunc(impl.helmAppRestHandler.GetDesiredManifest).Methods("POST")

	helmRouter.Path("/update").HandlerFunc(impl.helmAppRestHandler.UpdateApplication).Methods("PUT")

	helmRouter.Path("/delete").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.DeleteApplication).Methods("DELETE")

	helmRouter.Path("/template-chart").HandlerFunc(impl.helmAppRestHandler.TemplateChart).Methods("POST")
}
