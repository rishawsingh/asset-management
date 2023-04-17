package handler

import (
	"InternalAssetManagement/database"
	"InternalAssetManagement/database/dbhelper"
	"InternalAssetManagement/models"
	"InternalAssetManagement/utils"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jmoiron/sqlx"
)

func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var EmployeeDetails models.EmployeeDetails

	if parseErr := utils.ParseBody(r.Body, &EmployeeDetails); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "CreateEmployee: Failed to parse request body.")
		return
	}

	validationErr := validate.Struct(EmployeeDetails)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error")
		return
	}

	err := dbhelper.CreateEmployee(&EmployeeDetails)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "CreateEmployee: cannot create employee.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Created employee successfully.",
	})
}

func GetEmployeeMoreInfo(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeID")

	filterCheck, err := utils.Filters(r)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "GetEmployeeMoreInfo: cannot get filters properly: ")
		return
	}

	filterCheck.EmployeeID = employeeID

	employee, empErr := dbhelper.GetEmployee(&filterCheck)
	if empErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, empErr, "GetEmployeeMoreInfo: failed to get employee list.")
		return
	}

	assetHistory, assetErr := dbhelper.GetAssetHistory(employeeID)
	if assetErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, assetErr, "GetEmployeeMoreInfo:failed to get asset history.")
		return
	}

	employee.GetEmployee[0].AssetHistory = assetHistory

	utils.RespondJSON(w, http.StatusOK, employee)
}

func GetEmployeeList(w http.ResponseWriter, r *http.Request) {
	filterCheck, err := utils.Filters(r)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "GetEmployeeList: cannot get filters properly: ")
		return
	}

	employee, empErr := dbhelper.GetEmployee(&filterCheck)
	if empErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, empErr, "failed to get employee list.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, employee)
}

func UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	body := models.EmployeeDetails{}
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body.")
		return
	}

	validationErr := validate.Struct(body)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error")
		return
	}

	if body.Status == utils.NotAnEmployee {
		count, err := dbhelper.GetAssignedAsset(body.ID)
		switch {
		case err != nil && count < 0:
			utils.RespondError(w, http.StatusInternalServerError, err, "Cannot check if some asset is assigned to employee.")
			return
		case count > 0:
			utils.RespondError(w, http.StatusBadRequest, err, "Cannot Update to -> Not an employee: Asset is assigned to this employee.")
			return
		}
	}

	updateErr := dbhelper.UpdateEmployee(&body)
	if updateErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, updateErr, "failed to update user details.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Employee details updated.",
	})
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeID")

	var body models.Employee
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body.")
		return
	}

	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "cannot get user id.")
		return
	}

	validationErr := validate.Struct(body)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error.")
		return
	}

	count, err := dbhelper.GetAssignedAsset(employeeID)
	switch {
	case err != nil && count < 0:
		utils.RespondError(w, http.StatusInternalServerError, err, "Cannot check if some asset is assigned to employee.")
		return
	case count > 0:
		utils.RespondError(w, http.StatusBadRequest, err, "Cannot delete: Asset is assigned to this employee.")
		return
	}

	err = dbhelper.DeleteEmployee(employeeID, userID, body)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to delete employee.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Employee deleted successfully.",
	})
}

func CreateEmployeeAssetRelation(w http.ResponseWriter, r *http.Request) {
	var employeeAssetRelation models.EmployeeAssetRelation

	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "Cannot get user details.")
		return
	}

	if parseErr := utils.ParseBody(r.Body, &employeeAssetRelation); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "CreateEmployeeAssetRelation: Failed to parse request body.")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		err := dbhelper.CreateEmployeeAssetRelation(employeeAssetRelation, userID, tx)
		if err != nil {
			return err
		}

		err = dbhelper.UpdateAvailableAsset(employeeAssetRelation.AssetID, false, utils.Assigned, tx)
		if err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "CreateEmployeeAssetRelation: cannot create employee asset relation.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Asset assigned successfully.",
	})
}

func GetAssetHistory(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employeeId")

	assetHistory, assetErr := dbhelper.GetAssetHistory(employeeID)
	if assetErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, assetErr, "failed to get asset history.")
		return
	}
	if assetHistory == nil {
		utils.RespondJSON(w, http.StatusOK, []models.AssetHistory{})
		return
	}

	utils.RespondJSON(w, http.StatusOK, assetHistory)
}
