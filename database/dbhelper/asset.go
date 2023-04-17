package dbhelper

import (
	"InternalAssetManagement/database"
	"InternalAssetManagement/models"
	"InternalAssetManagement/utils"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func CreateAsset(db *sqlx.Tx, assetDetails *models.CreateAsset, userID string) (string, error) {
	SQL := `INSERT INTO assets (brand, model, serial_no, asset_type, purchased_date, warranty_start_date, warranty_expiry_date,
								created_by, owned_by, client_name)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id`
	var id string
	err := db.Get(&id, SQL, assetDetails.Brand, assetDetails.Model, assetDetails.SerialNo, assetDetails.AssetType, assetDetails.PurchasedDate, assetDetails.WarrantyStartDate, assetDetails.WarrantyExpiryDate, userID, assetDetails.OwnedBy, assetDetails.ClientName)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("CreateAsset: cannot create asset.")
		return "", err
	}
	return id, nil
}

func CreateLaptopSpecification(db *sqlx.Tx, assetDetails *models.CreateAsset, assetID string) error {
	SQL := `INSERT INTO laptop_specifications (asset_id, series, processor, ram, operating_system, charger, screen_resolution,
											   storage)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.Exec(SQL, assetID, assetDetails.Series, assetDetails.Processor, assetDetails.RAM, assetDetails.OperatingSystem, assetDetails.Charger, assetDetails.ScreenResolution, assetDetails.Storage)
	if err != nil {
		logrus.WithError(err).Error("CreateLaptopSpecification: cannot create laptop specification.")
		return err
	}
	return nil
}

func CreatePenDriveSpecification(db *sqlx.Tx, assetDetails *models.CreateAsset, assetID string) error {
	SQL := `INSERT INTO pen_drive_specifications (asset_id, storage)
			VALUES ($1, $2)`
	_, err := db.Exec(SQL, assetID, assetDetails.Storage)
	if err != nil {
		logrus.WithError(err).Error("CreatePenDriveSpecification: cannot create pen drive specification.")
		return err
	}
	return nil
}

func CreateHardDiskSpecification(db *sqlx.Tx, assetDetails *models.CreateAsset, assetID string) error {
	SQL := `INSERT INTO hard_disk_specifications (asset_id, storage)
			VALUES ($1,$2)`
	_, err := db.Exec(SQL, assetID, assetDetails.Storage)
	if err != nil {
		logrus.WithError(err).Error("CreateHardDiskSpecification: cannot create hard disk specification.")
		return err
	}
	return nil
}

func CreateMobileSpecification(db *sqlx.Tx, assetDetails *models.CreateAsset, assetID string) error {
	SQL := `INSERT INTO mobile_specifications (asset_id, os_type, imei_1, imei_2, ram)
			VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(SQL, assetID, assetDetails.OsType, assetDetails.Imei1, assetDetails.Imei2, assetDetails.RAM)
	if err != nil {
		logrus.WithError(err).Error("CreateMobileSpecification: cannot create mobile specification.")
		return err
	}
	return nil
}

func CreateSimSpecification(db *sqlx.Tx, assetDetails *models.CreateAsset, assetID string) error {
	SQL := `INSERT INTO sim_specifications (asset_id, sim_no, phone_no)
			VALUES ($1, $2, $3)`
	_, err := db.Exec(SQL, assetID, assetDetails.SimNo, assetDetails.PhoneNo)
	if err != nil {
		logrus.WithError(err).Error("CreateSimSpecification: cannot create sim specification.")
		return err
	}
	return nil
}

func GetAssetSpec(assetID, assetType string) ([]models.CreateAsset, error) {
	SQL := `SELECT 
    				brand, 
    				model, 
    				serial_no, 
    				asset_type,
       				purchased_date,
       				status,
       				warranty_expiry_date,
       				assets.archived_at,
       				archive_reason,
       				deleted_by,
       				owned_by,
       				client_name
          `
	num := 1
	values := make([]interface{}, 0)
	switch assetType {
	case utils.Laptop:
		assetStr := fmt.Sprintf(",series,processor,ram,operating_system,charger,screen_resolution,storage FROM  assets LEFT JOIN laptop_specifications ls on assets.id = ls.asset_id WHERE asset_id = $%d", num)
		SQL += assetStr
		values = append(values, assetID)
	case utils.Pendrive:
		assetStr := fmt.Sprintf(", storage FROM  assets LEFT JOIN pen_drive_specifications ls on assets.id = ls.asset_id WHERE asset_id = $%d", num)
		SQL += assetStr
		values = append(values, assetID)
	case utils.Mouse:
		assetStr := fmt.Sprintf("FROM  assets  WHERE id = $%d", num)
		SQL += assetStr
		values = append(values, assetID)
	case utils.Harddisk:
		assetStr := fmt.Sprintf(", storage FROM  assets LEFT JOIN hard_disk_specifications ls on assets.id = ls.asset_id WHERE asset_id = $%d", num)
		SQL += assetStr
		values = append(values, assetID)
	case utils.Mobile:
		assetStr := fmt.Sprintf(", os_type, imei_1, imei_2, ram FROM  assets LEFT JOIN mobile_specifications ls on assets.id = ls.asset_id WHERE asset_id = $%d", num)
		SQL += assetStr
		values = append(values, assetID)
	case utils.Sim:
		assetStr := fmt.Sprintf(", sim_no, phone_no FROM  assets LEFT JOIN sim_specifications ls on assets.id = ls.asset_id WHERE asset_id = $%d", num)
		SQL += assetStr
		values = append(values, assetID)
	}
	var assetSpec = make([]models.CreateAsset, 0)
	err := database.AssetManagement.Select(&assetSpec, SQL, values...)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("GetAssetSpec: cannot get asset specifications.")
		return assetSpec, err
	}
	return assetSpec, nil
}

func GetAssetsWithFilters(filterCheck *models.
	FiltersCheck) (models.TotalGetAsset, error) {
	var totalGetAsset models.TotalGetAsset
	SQL := `WITH cte_asset AS(  SELECT  count(*) over () as total_count,
                        				a.id,
        								brand,
        								model,
        								serial_no,
        								asset_type,
        								purchased_date,
        								a.status,
        								warranty_expiry_date,
                                        case when a.status = 'assigned' then e.id else null end as assigned_to_id,
                                        case when a.status = 'assigned' OR a.status = 'deleted' then e.name else '' end as name
								FROM assets a LEFT JOIN employee_asset_relation ear on a.id = ear.asset_id
												   LEFT JOIN employee e on e.id = ear.employee_id
								WHERE 
  			`
	values := make([]interface{}, 0)
	args := 0

	if filterCheck.Available {
		filterStr := fmt.Sprintf("a.status = $%d ", args+1)
		SQL += filterStr
		args++
		values = append(values, utils.Available)
	}
	if filterCheck.Assigned {
		var filterStr string
		if !filterCheck.Available {
			filterStr = fmt.Sprintf("a.status = $%d ", args+1)
		} else {
			filterStr = fmt.Sprintf("OR a.status = $%d ", args+1)
		}
		SQL += filterStr
		args++
		values = append(values, utils.Assigned)
	}
	if filterCheck.Deleted {
		var filterStr string
		if !filterCheck.Available && !filterCheck.Assigned {
			filterStr = fmt.Sprintf("a.status = $%d ", args+1)
		} else {
			filterStr = fmt.Sprintf("OR a.status = $%d ", args+1)
		}
		SQL += filterStr
		args++
		values = append(values, utils.Deleted)
	}

	if len(filterCheck.AssetTypes) > 0 {
		assetStr := fmt.Sprintf("AND a.asset_type =ANY($%d)", args+1)
		SQL += assetStr
		args++
		values = append(values, pq.Array(filterCheck.AssetTypes))
	}

	if filterCheck.SearchedName != "" {
		nameStr := fmt.Sprintf("AND (brand ilike '%%' || $%d || '%%')", args+1)
		SQL += nameStr
		args++
		values = append(values, filterCheck.SearchedName)
	}

	if filterCheck.Warranty > 0 {
		warrantyStr := fmt.Sprintf("AND a.warranty_expiry_date BETWEEN now() and (now() + ($%d ||' months')::interval)", args+1)
		SQL += warrantyStr
		args++
		values = append(values, filterCheck.Warranty)
	} else if filterCheck.Warranty == 0 && filterCheck.IsExpired {
		warrantyStr := fmt.Sprintf("AND a.warranty_expiry_date < $%d ", args+1)
		SQL += warrantyStr
		args++
		values = append(values, time.Now())
	}

	if filterCheck.Pagination {
		//nolint:gomnd // addition of constant
		pageStr := fmt.Sprintf("ORDER BY id LIMIT $%d OFFSET $%d)SELECT total_count,id,brand,model,serial_no,asset_type,purchased_date,status,warranty_expiry_date, assigned_to_id, name FROM cte_asset", args+1, args+2)
		SQL += pageStr
		values = append(values, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)
	} else {
		countStr := `)SELECT id,brand,model,serial_no,asset_type,purchased_date,status,warranty_expiry_date FROM cte_asset`
		SQL += countStr
	}

	var assets = make([]models.GetAsset, 0)
	err := database.AssetManagement.Select(&assets, SQL, values...)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("GetAvailableAssets: cannot get available assets.")
		return totalGetAsset, err
	}
	if len(assets) == 0 {
		totalGetAsset.GetAsset = []models.GetAsset{}
		logrus.WithError(err).Error("GetAvailableAssets: got empty array.")
		return totalGetAsset, err
	}

	totalGetAsset.GetAsset = assets
	totalGetAsset.TotalCount = assets[0].TotalCount
	return totalGetAsset, nil
}

func GetAssets(filterCheck *models.FiltersCheck) (models.TotalGetAsset, error) {
	var totalGetAsset models.TotalGetAsset
	// language = sql
	SQL := `with cte_asset AS(select distinct on(a.id)
                              a.id,
                              brand,
                              model,
                              serial_no,
                              asset_type,
                              purchased_date,
                              a.status,
                              warranty_expiry_date,
                              case when is_available = false then e.id else null end as assigned_to_id,
                              case when is_available = false then e.name else '' end as name
                  FROM assets a
                           LEFT JOIN employee_asset_relation ear on a.id = ear.asset_id
                           LEFT JOIN employee e on ear.employee_id = e.id
                  WHERE a.archived_at IS NULL
                    AND e.archived_at IS NULL
`
	values := make([]interface{}, 0)
	args := 0

	if len(filterCheck.AssetTypes) > 0 {
		assetStr := fmt.Sprintf("AND asset_type =ANY($%d)", args+1)
		SQL += assetStr
		args++
		values = append(values, pq.Array(filterCheck.AssetTypes))
	}

	if filterCheck.SearchedName != "" {
		nameStr := fmt.Sprintf("AND (brand ilike '%%' || $%d || '%%')", args+1)
		SQL += nameStr
		args++
		values = append(values, filterCheck.SearchedName)
	}

	if filterCheck.Warranty > 0 {
		warrantyStr := fmt.Sprintf("AND warranty_expiry_date BETWEEN now() and (now() + ($%d ||' months')::interval)", args+1)
		SQL += warrantyStr
		args++
		values = append(values, filterCheck.Warranty)
	} else if filterCheck.Warranty == 0 && filterCheck.IsExpired {
		warrantyStr := fmt.Sprintf("AND warranty_expiry_date < $%d ", args+1)
		SQL += warrantyStr
		args++
		values = append(values, time.Now())
	}
	//nolint:gomnd // addition of constant
	pageStr := fmt.Sprintf("ORDER BY a.id, ear.retrieved_date DESC)SELECT count(*) over() as total_count, id,  brand, model, serial_no, asset_type, purchased_date, status, warranty_expiry_date, assigned_to_id, name FROM  cte_asset LIMIT $%d OFFSET $%d", args+1, args+2)
	SQL += pageStr
	values = append(values, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)

	var assets = make([]models.GetAsset, 0)
	err := database.AssetManagement.Select(&assets, SQL, values...)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("GetAssets: cannot get assets.")
		return totalGetAsset, err
	}

	if len(assets) == 0 {
		totalGetAsset.GetAsset = []models.GetAsset{}
		logrus.WithError(err).Error("GetAssets: got empty array.")
		return totalGetAsset, err
	}

	totalGetAsset.GetAsset = assets
	totalGetAsset.TotalCount = assets[0].TotalCount
	return totalGetAsset, nil
}

func UpdateAsset(assetDetails *models.UpdateAssetSpecification, tx *sqlx.Tx) error {
	SQL := `UPDATE assets
			SET brand                = $1,
				model                = $2,
				serial_no            = $3,
				purchased_date       = $4,
				warranty_start_date  = $5,
				warranty_expiry_date = $6,
				updated_at           = NOW()
			WHERE id = $7
			  AND archived_at IS NULL`
	_, err := tx.Exec(SQL, assetDetails.Brand, assetDetails.Model, assetDetails.SerialNo, assetDetails.PurchasedDate, assetDetails.WarrantyStartDate, assetDetails.WarrantyExpiryDate, assetDetails.ID)
	if err != nil {
		logrus.WithError(err).Error("UpdateAsset: cannot update asset.")
		return err
	}
	return nil
}

func UpdateLaptopSpecifications(assetSpecifications *models.UpdateAssetSpecification, tx *sqlx.Tx) error {
	SQL := `UPDATE laptop_specifications
			SET series            = $1,
				processor         = $2,
				ram               = $3,
				operating_system  = $4,
				charger           = $5,
				screen_resolution = $6,
				storage           = $7,
				updated_at        = NOW()
			WHERE asset_id = $8
			  AND archived_at IS NULL`
	_, err := tx.Exec(SQL, assetSpecifications.Series, assetSpecifications.Processor, assetSpecifications.RAM, assetSpecifications.OperatingSystem, assetSpecifications.Charger, assetSpecifications.ScreenResolution, assetSpecifications.Storage, assetSpecifications.ID)
	if err != nil {
		logrus.WithError(err).Error("UpdateLaptopSpecifications: cannot update laptop specifications.")
		return err
	}
	return nil
}

func UpdateHardDiskSpecifications(storage, id string, tx *sqlx.Tx) error {
	SQL := `UPDATE hard_disk_specifications
			SET storage = $1
			WHERE asset_id = $2
			  AND archived_at IS NULL`
	_, err := tx.Exec(SQL, storage, id)
	if err != nil {
		logrus.WithError(err).Error("UpdateHardDiskSpecifications: cannot update hard disk specifications.")
		return err
	}
	return nil
}

func UpdatePenDriveSpecifications(storage, id string, tx *sqlx.Tx) error {
	SQL := `UPDATE pen_drive_specifications
			SET storage = $1
			WHERE asset_id = $2
			  AND archived_at IS NULL`
	_, err := tx.Exec(SQL, storage, id)
	if err != nil {
		logrus.WithError(err).Error("UpdatePenDriveSpecifications: cannot update pen drive specifications.")
		return err
	}
	return nil
}

func UpdateMobileSpecifications(assetSpecifications *models.UpdateAssetSpecification, tx *sqlx.Tx) error {
	SQL := `UPDATE mobile_specifications
            SET    os_type = $1,
                   imei_1 = $2,
                   imei_2 = $3,
                   ram = $4,
                   updated_at = now()
            WHERE  asset_id = $5
            AND    archived_at IS NULL 
            `
	_, err := tx.Exec(SQL, assetSpecifications.OsType, assetSpecifications.Imei1, assetSpecifications.Imei2, assetSpecifications.RAM, assetSpecifications.ID)
	if err != nil {
		logrus.WithError(err).Error("UpdateMobileSpecifications: cannot update mobile specifications.")
		return err
	}
	return nil
}

func UpdateSimSpecifications(assetSpecifications *models.UpdateAssetSpecification, tx *sqlx.Tx) error {
	SQL := `UPDATE sim_specifications
            SET    sim_no = $1,
                   phone_no = $2,
                   updated_at = now()
            WHERE  asset_id = $3
            AND    archived_at IS NULL 
            `
	_, err := tx.Exec(SQL, assetSpecifications.SimNo, assetSpecifications.PhoneNo, assetSpecifications.ID)
	if err != nil {
		logrus.WithError(err).Error("UpdateSimSpecifications: cannot update sim specifications.")
		return err
	}
	return nil
}

func RetrieveAssetByAssetID(db *sqlx.Tx, retrievalDetails *models.ReassignAsset) error {
	SQL := `UPDATE employee_asset_relation
            SET    retrieved_date = $1,
                   retrieval_reason = $2,
                   archived_at = NOW()
            WHERE  asset_id = $3
            AND    archived_at IS NULL 
            `
	_, err := db.Exec(SQL, retrievalDetails.RetrievedDate, retrievalDetails.RetrievalReason, retrievalDetails.AssetID)
	if err != nil {
		logrus.WithError(err).Error("RetrieveAssetByAssetID: cannot retrieve asset.")
		return err
	}
	return nil
}

func UpdateAvailableAsset(assetID string, availableBool bool, status string, tx *sqlx.Tx) error {
	SQL := `UPDATE assets
            SET    is_available = $1,
                   status = $3
            WHERE  id = $2
            AND archived_at IS NULL 
            `

	_, err := tx.Exec(SQL, availableBool, assetID, status)
	if err != nil {
		logrus.WithError(err).Error("UpdateAvailableAsset: cannot update available asset.")
		return err
	}
	return nil
}

func ReassignAsset(db *sqlx.Tx, reassignDetails *models.ReassignAsset, assignedBy string) error {
	SQL := `INSERT INTO employee_asset_relation(employee_id, asset_id, assigned_by, assigned_date)
            VALUES ($1, $2, $3, $4)`

	_, err := db.Exec(SQL, reassignDetails.EmployeeID, reassignDetails.AssetID, assignedBy, reassignDetails.AssignedDate)
	if err != nil {
		logrus.WithError(err).Error("ReassignAsset: unable to reassign asset.")
		return err
	}
	return nil
}

func AvailableAssets(brand, assetType, modelNo string) ([]models.AssignAssetDetails, error) {
	SQL := `SELECT `
	values := make([]interface{}, 0)
	args := 1
	if assetType != "" && brand == "" {
		assetStr := fmt.Sprintf("DISTINCT ON (brand) brand FROM assets WHERE asset_type = $%d AND archived_at IS NULL AND is_available = true", args)
		SQL += assetStr
		values = append(values, assetType)
	} else if brand != "" {
		switch {
		case assetType == utils.Sim:
			assetStr := fmt.Sprintf("a.id, sim_no FROM assets a LEFT JOIN sim_specifications ss ON a.id = ss.asset_id  WHERE brand = $%d AND a.archived_at IS NULL AND a.is_available = true AND ss.archived_at IS NULL", args)
			SQL += assetStr
			values = append(values, brand)
		case modelNo == "":
			//nolint:gomnd // constant value
			assetStr := fmt.Sprintf("model FROM assets WHERE brand = $%d AND asset_type = $%d AND archived_at IS NULL AND is_available = true", args, 2)
			SQL += assetStr
			values = append(values, brand, assetType)
		case assetType == utils.Mobile:
			//nolint:gomnd // constant value
			assetStr := fmt.Sprintf("a.id, imei_1 FROM assets a LEFT JOIN mobile_specifications ms ON a.id = ms.asset_id WHERE brand = $%d AND model = $%d AND a.archived_at IS NULL AND a.is_available = true AND ms.archived_at IS NULL", args, 2)
			SQL += assetStr
			values = append(values, brand, modelNo)
		default:
			//nolint:gomnd // constant value
			assetStr := fmt.Sprintf("id, serial_no FROM assets WHERE brand = $%d AND model = $%d AND archived_at IS NULL AND is_available = true", args, 2)
			SQL += assetStr
			values = append(values, brand, modelNo)
		}
	}

	brandName := make([]models.AssignAssetDetails, 0)
	err := database.AssetManagement.Select(&brandName, SQL, values...)
	if err != nil {
		logrus.WithError(err).Error("AvailableAssets: cannot get assigned asset details.")
		return brandName, err
	}
	return brandName, nil
}

func EmployeeHistory(assetID string) ([]models.EmployeeHistory, error) {
	SQL := `SELECT  e.id,
       				name,
       				email,
       				phone_no,
       				assigned_by,
       				assigned_date,
       				retrieved_date,
       				COALESCE(retrieval_reason, '') as retrieval_reason
			FROM   employee e 
			    JOIN employee_asset_relation ear on e.id = ear.employee_id
			WHERE  e.archived_at IS NULL
			AND    asset_id = $1`

	employeeHistory := make([]models.EmployeeHistory, 0)
	err := database.AssetManagement.Select(&employeeHistory, SQL, assetID)
	if err != nil {
		logrus.WithError(err).Error("EmployeeHistory: cannot get employee history.")
		return employeeHistory, err
	}
	return employeeHistory, nil
}

func UpdateWarranty(warrantyDetails models.WarrantyDetails) error {
	SQL := `UPDATE assets
            SET    warranty_start_date = $1,
                   warranty_expiry_date = $2
            WHERE archived_at IS NULL 
            AND   id = $3`
	_, err := database.AssetManagement.Exec(SQL, warrantyDetails.WarrantyStartDate, warrantyDetails.WarrantyExpiryDate, warrantyDetails.AssetID)
	if err != nil {
		logrus.WithError(err).Error("UpdateWarranty: cannot update warranty details.")
		return err
	}
	return nil
}

func RetrieveAsset(assetRetrievalDetails models.AssetRetrievalDetails, tx *sqlx.Tx) error {
	SQL := `UPDATE employee_asset_relation
            SET    retrieved_date = $1,
                   retrieval_reason = $2
            WHERE  employee_id = $3
            AND    asset_id = $4
            AND    retrieved_date IS NULL 
           `
	_, err := tx.Exec(SQL, assetRetrievalDetails.RetrievedDate, assetRetrievalDetails.RetrievalReason, assetRetrievalDetails.EmployeeID, assetRetrievalDetails.AssetID)
	if err != nil {
		logrus.WithError(err).Error("RetrieveAsset: cannot update retrieval details.")
		return err
	}
	return nil
}

func GetAssignedEmployee(asset models.Asset) (int, error) {
	SQL := `SELECT COUNT(id)
			FROM employee_asset_relation
			WHERE asset_id = $1
			  AND retrieved_date IS NULL
			  AND archived_at IS NULL`
	var count int
	err := database.AssetManagement.Get(&count, SQL, asset.ID)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("GetAssignedEmployee: cannot get assigned employee.")
		return -1, err
	}
	return count, nil
}

func GetAssignedAsset(employeeID string) (int, error) {
	SQL := `SELECT COUNT(id)
			FROM employee_asset_relation
			WHERE employee_id = $1
			  AND retrieved_date IS NULL
			  AND archived_at IS NULL`
	var count int
	err := database.AssetManagement.Get(&count, SQL, employeeID)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("GetAssignedAsset: cannot get assigned asset.")
		return -1, err
	}
	return count, nil
}

func DeleteLaptopSpec(db *sqlx.Tx, id string) error {
	SQL := `UPDATE laptop_specifications
			SET archived_at = NOW()
			WHERE asset_id = $1
			  AND archived_at IS NULL`
	_, err := db.Exec(SQL, id)
	if err != nil {
		logrus.WithError(err).Error("DeleteLaptopSpec: cannot delete laptop specifications.")
		return err
	}
	return nil
}

func DeletePenDriveSpec(db *sqlx.Tx, id string) error {
	SQL := `UPDATE pen_drive_specifications
			SET archived_at = NOW()
			WHERE asset_id = $1
			  AND archived_at IS NULL`
	_, err := db.Exec(SQL, id)
	if err != nil {
		logrus.WithError(err).Error("DeletePenDriveSpec: cannot delete pen drive specifications.")
		return err
	}
	return nil
}

func DeleteHardDiskSpec(db *sqlx.Tx, id string) error {
	SQL := `UPDATE hard_disk_specifications
			SET archived_at = NOW()
			WHERE asset_id = $1
			  AND archived_at IS NULL`
	_, err := db.Exec(SQL, id)
	if err != nil {
		logrus.WithError(err).Error("DeleteHardDiskSpec: cannot delete hard disk specifications.")
		return err
	}
	return nil
}

func DeleteMobileSpec(db *sqlx.Tx, id string) error {
	SQL := `UPDATE mobile_specifications
			SET archived_at = NOW()
			WHERE asset_id = $1
			  AND archived_at IS NULL`
	_, err := db.Exec(SQL, id)
	if err != nil {
		logrus.WithError(err).Error("DeleteMobileSpec: cannot delete mobile specifications.")
		return err
	}
	return nil
}

func DeleteSimSpec(db *sqlx.Tx, id string) error {
	SQL := `UPDATE sim_specifications
			SET archived_at = NOW()
			WHERE asset_id = $1
			  AND archived_at IS NULL`
	_, err := db.Exec(SQL, id)
	if err != nil {
		logrus.WithError(err).Error("DeleteSimSpec: cannot delete sim specifications.")
		return err
	}
	return nil
}

func DeleteAsset(db *sqlx.Tx, assetDetails models.Asset, userID string) error {
	SQL := `UPDATE assets
			SET archived_at = NOW(),
			    status = $2,
			    archive_reason = $3,
			    deleted_by = $4
			WHERE id = $1
			  AND archived_at IS NULL`
	_, err := db.Exec(SQL, assetDetails.ID, utils.Deleted, assetDetails.DeleteReason, userID)
	if err != nil {
		logrus.WithError(err).Error("DeleteAsset: cannot delete asset.")
		return err
	}
	return nil
}
