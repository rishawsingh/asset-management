package handler

import (
	"InternalAssetManagement/database"
	"InternalAssetManagement/database/dbhelper"
	"InternalAssetManagement/models"
	"InternalAssetManagement/utils"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func CreateAsset(w http.ResponseWriter, r *http.Request) {
	body := models.CreateAsset{}
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body.")
		return
	}

	validationErr := validate.Struct(body)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error")
		return
	}

	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "cannot get user details.")
		return
	}

	if body.OwnedBy == utils.RemoteState {
		body.ClientName = ""
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		assetID, assetErr := dbhelper.CreateAsset(tx, &body, userID)
		if assetErr != nil {
			return assetErr
		}

		switch body.AssetType {
		case models.Laptop:
			err := dbhelper.CreateLaptopSpecification(tx, &body, assetID)
			if err != nil {
				return err
			}
		case models.Pendrive:
			err := dbhelper.CreatePenDriveSpecification(tx, &body, assetID)
			if err != nil {
				return err
			}
		case models.Harddisk:
			err := dbhelper.CreateHardDiskSpecification(tx, &body, assetID)
			if err != nil {
				return err
			}
		case models.Mobile:
			err := dbhelper.CreateMobileSpecification(tx, &body, assetID)
			if err != nil {
				return err
			}
		case models.Sim:
			err := dbhelper.CreateSimSpecification(tx, &body, assetID)
			if err != nil {
				return err
			}
		case models.Mouse:
			break
		}

		return nil
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "failed to create asset.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Asset created.",
	})
}

func GetAssetSpec(w http.ResponseWriter, r *http.Request) {
	assetID := r.URL.Query().Get("assetId")
	assetType := r.URL.Query().Get("assetType")

	assetSpec, err := dbhelper.GetAssetSpec(assetID, assetType)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "GetAssetSpec: cannot asset spec.")
		return
	}

	employeeHistory, err := dbhelper.EmployeeHistory(assetID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "EmployeeHistory: cannot get employee history.")
		return
	}

	assetSpec[0].AssetHistory = employeeHistory

	if assetSpec == nil {
		utils.RespondJSON(w, http.StatusOK, []models.CreateAsset{})
		return
	}

	utils.RespondJSON(w, http.StatusOK, assetSpec)
}

func GetAssetList(w http.ResponseWriter, r *http.Request) {
	filterCheck, err := utils.Filters(r)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "GetAssetLIst: cannot get filters properly: ")
		return
	}
	var assets models.TotalGetAsset
	var assetErr error
	switch {
	case filterCheck.Available || filterCheck.Assigned || filterCheck.Deleted:
		assets, assetErr = dbhelper.GetAssetsWithFilters(&filterCheck)
	default:
		assets, assetErr = dbhelper.GetAssets(&filterCheck)
	}
	if assetErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, assetErr, "Failed to get Asset List.")
		return
	}
	utils.RespondJSON(w, http.StatusOK, assets)
}

func UpdateAsset(w http.ResponseWriter, r *http.Request) {
	body := models.UpdateAssetSpecification{}
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "Failed to parse request body.")
		return
	}

	validationErr := validate.Struct(body)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		assetErr := dbhelper.UpdateAsset(&body, tx)
		if assetErr != nil {
			return assetErr
		}

		switch body.AssetType {
		case models.Laptop:
			updateErr := dbhelper.UpdateLaptopSpecifications(&body, tx)
			if updateErr != nil {
				return updateErr
			}
		case models.Harddisk:
			updateErr := dbhelper.UpdateHardDiskSpecifications(body.Storage, body.ID, tx)
			if updateErr != nil {
				return updateErr
			}
		case models.Pendrive:
			updateErr := dbhelper.UpdatePenDriveSpecifications(body.Storage, body.ID, tx)
			if updateErr != nil {
				return updateErr
			}
		case models.Mobile:
			updateErr := dbhelper.UpdateMobileSpecifications(&body, tx)
			if updateErr != nil {
				return updateErr
			}
		case models.Sim:
			updateErr := dbhelper.UpdateSimSpecifications(&body, tx)
			if updateErr != nil {
				return updateErr
			}
		case models.Mouse:
			break
		}

		return nil
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "failed to update asset.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Asset updated successfully.",
	})
}

func ReassignAsset(w http.ResponseWriter, r *http.Request) {
	body := models.ReassignAsset{}
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body.")
		return
	}

	validationErr := validate.Struct(body)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error")
		return
	}

	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "cannot get user details.")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		retrieveErr := dbhelper.RetrieveAssetByAssetID(tx, &body)
		if retrieveErr != nil {
			return retrieveErr
		}

		err := dbhelper.UpdateAvailableAsset(body.AssetID, false, utils.Assigned, tx)
		if err != nil {
			return err
		}

		reassignErr := dbhelper.ReassignAsset(tx, &body, userID)
		if reassignErr != nil {
			return reassignErr
		}

		return nil
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "failed to re-assign asset.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Asset re-assigned successfully.",
	})
}

func AvailableAssets(w http.ResponseWriter, r *http.Request) {
	assetType := r.URL.Query().Get("assetType")
	brand := r.URL.Query().Get("brand")
	modelNo := r.URL.Query().Get("model")
	assets, err := dbhelper.AvailableAssets(brand, assetType, modelNo)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "AvailableAssets: cannot get assigned asset details.")
		return
	}
	if assets == nil {
		utils.RespondJSON(w, http.StatusOK, []string{})
		return
	}

	utils.RespondJSON(w, http.StatusOK, assets)
}

func EmployeeHistory(w http.ResponseWriter, r *http.Request) {
	assetID := r.URL.Query().Get("assetID")

	employeeHistory, err := dbhelper.EmployeeHistory(assetID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "EmployeeHistory: cannot get employee history.")
		return
	}
	if employeeHistory == nil {
		utils.RespondJSON(w, http.StatusOK, []models.EmployeeHistory{})
		return
	}

	utils.RespondJSON(w, http.StatusOK, employeeHistory)
}

func UpdateWarranty(w http.ResponseWriter, r *http.Request) {
	var warrantyDetails models.WarrantyDetails

	if parseErr := utils.ParseBody(r.Body, &warrantyDetails); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "Failed to parse request body.")
		return
	}

	err := dbhelper.UpdateWarranty(warrantyDetails)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "UpdateWarranty: cannot update warranty.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Warranty updated successfully.",
	})
}

func RetrieveAsset(w http.ResponseWriter, r *http.Request) {
	var assetRetrievalDetails models.AssetRetrievalDetails

	if parseErr := utils.ParseBody(r.Body, &assetRetrievalDetails); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "Failed to parse request body.")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		err := dbhelper.RetrieveAsset(assetRetrievalDetails, tx)
		if err != nil {
			return err
		}

		err = dbhelper.UpdateAvailableAsset(assetRetrievalDetails.AssetID, true, utils.Available, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "RetrieveAsset: cannot update retrieval details.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Asset retrieved successfully.",
	})
}

func DeleteAsset(w http.ResponseWriter, r *http.Request) {
	body := models.Asset{}
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

	count, err := dbhelper.GetAssignedEmployee(body)
	switch {
	case err != nil && count < 0:
		utils.RespondError(w, http.StatusInternalServerError, err, "cannot check if asset is assigned.")
		return
	case count > 0:
		utils.RespondError(w, http.StatusBadRequest, err, "Asset is assigned to someone.")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		switch {
		case body.AssetType == models.Laptop:
			err := dbhelper.DeleteLaptopSpec(tx, body.ID)
			if err != nil {
				return err
			}
		case body.AssetType == models.Pendrive:
			err := dbhelper.DeletePenDriveSpec(tx, body.ID)
			if err != nil {
				return err
			}
		case body.AssetType == models.Harddisk:
			err := dbhelper.DeleteHardDiskSpec(tx, body.ID)
			if err != nil {
				return err
			}
		case body.AssetType == models.Mobile:
			err := dbhelper.DeleteMobileSpec(tx, body.ID)
			if err != nil {
				return err
			}
		case body.AssetType == models.Sim:
			err := dbhelper.DeleteSimSpec(tx, body.ID)
			if err != nil {
				return err
			}
		}

		deleteErr := dbhelper.DeleteAsset(tx, body, userID)
		if deleteErr != nil {
			return deleteErr
		}

		return nil
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "failed to delete asset.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Asset deleted successfully.",
	})
}
