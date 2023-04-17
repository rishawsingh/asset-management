package server

import (
	"InternalAssetManagement/handler"

	"github.com/go-chi/chi/v5"
)

func employeeRoutes(r chi.Router) {
	r.Group(func(employee chi.Router) {
		employee.Post("/", handler.CreateEmployee)
		employee.Get("/", handler.GetEmployeeList)
		employee.Put("/", handler.UpdateEmployee)
		employee.Delete("/{employeeID}", handler.DeleteEmployee)

		employee.Get("/{employeeID}/info", handler.GetEmployeeMoreInfo)
		employee.Post("/asset", handler.CreateEmployeeAssetRelation)
		employee.Get("/asset-list", handler.GetAssetHistory)
	})
}
