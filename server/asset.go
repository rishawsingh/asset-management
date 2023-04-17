package server

import (
	"InternalAssetManagement/handler"

	"github.com/go-chi/chi/v5"
)

func assetRoutes(r chi.Router) {
	r.Group(func(asset chi.Router) {
		asset.Post("/", handler.CreateAsset)
		asset.Get("/specifications", handler.GetAssetSpec)
		asset.Get("/", handler.GetAssetList)
		asset.Put("/", handler.UpdateAsset)
		asset.Post("/reassign", handler.ReassignAsset)
		asset.Get("/brand", handler.AvailableAssets)

		asset.Get("/employee", handler.EmployeeHistory)
		asset.Put("/warranty", handler.UpdateWarranty)
		asset.Put("/retrieve-asset", handler.RetrieveAsset)
		asset.Delete("/", handler.DeleteAsset)
	})
}
