package dbhelper

import (
	"InternalAssetManagement/database"
	"InternalAssetManagement/models"
	"InternalAssetManagement/utils"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"

	"github.com/sirupsen/logrus"
)

func CreateEmployee(employeeDetails *models.EmployeeDetails) error {
	SQL := `INSERT INTO employee(name, email, phone_no, type) 
            VALUES($1, $2, $3, $4)
            ON CONFLICT (email) DO UPDATE 
            SET email = $2`

	_, err := database.AssetManagement.Exec(SQL, employeeDetails.Name, employeeDetails.Email, employeeDetails.PhoneNo, employeeDetails.Type)
	if err != nil {
		logrus.WithError(err).Error("CreateEmployee: cannot create employee.")
		return err
	}
	return nil
}

func GetEmployee(filterCheck *models.FiltersCheck) (models.TotalGetEmployee, error) {
	var totalGetEmployee models.TotalGetEmployee
	SQL := `WITH cte_employee AS (SELECT count(*) over () as total_count,
                             e.id             as id,
                             name,
                             email,
                             phone_no,
                             e.status,
                             e.type,
                             e.archived_at,
                             e.archive_reason,
                             e.deleted_by,
                             COUNT(ear.id)    AS asset_quantity
                      FROM employee e
                               LEFT JOIN employee_asset_relation ear ON e.id = ear.employee_id
                               LEFT JOIN assets a on a.id = ear.asset_id
                      WHERE e.id IS NOT NULL
							
`

	values := make([]interface{}, 0)
	args := 0

	sqlStr := fmt.Sprintf(" AND (LENGTH($%d) != 0 OR $%d OR e.archived_at IS NULL)   AND (CARDINALITY(Array[$%d::asset_type[]]) = 0 OR ( a.asset_type =ANY(ARRAY [$%d::asset_type[]])) ) AND ( NULLIF(LENGTH($%d), 0) IS NULL OR (name ilike '%%' || $%d || '%%') AND (NULLIF(LENGTH($%d), 0)) IS NULL OR e.id::text = $%d) ", args+1, args+2, args+3, args+4, args+5, args+6, args+7, args+8)
	SQL += sqlStr
	args += 8
	values = append(values, filterCheck.EmployeeID, filterCheck.Deleted, filterCheck.AssetTypes, filterCheck.AssetTypes, filterCheck.SearchedName, filterCheck.SearchedName, filterCheck.EmployeeID, filterCheck.EmployeeID)

	if filterCheck.Deleted && !filterCheck.NotAnEmployee {
		statusStr := fmt.Sprintf("AND e.status = $%d ", args+1)
		SQL += statusStr
		args++
		values = append(values, utils.Deleted)
	}
	if !filterCheck.Deleted && filterCheck.NotAnEmployee {
		statusStr := fmt.Sprintf("AND e.status = $%d ", args+1)
		SQL += statusStr
		args++
		values = append(values, utils.NotAnEmployee)
	}
	if filterCheck.Deleted && filterCheck.NotAnEmployee {
		statusStr := fmt.Sprintf("AND e.status = $%d OR e.status = $%d ", args+1, args+2)
		SQL += statusStr
		args += 2
		values = append(values, utils.Deleted, utils.NotAnEmployee)
	}
	if !filterCheck.Deleted && !filterCheck.NotAnEmployee {
		statusStr := fmt.Sprintf("AND e.status = $%d ", args+1)
		SQL += statusStr
		args++
		values = append(values, utils.Active)
	}

	pageStr := fmt.Sprintf(" GROUP BY (e.id, name, email, phone_no, e.status, e.type, e.archived_at, e.archive_reason, e.deleted_by) LIMIT $%d OFFSET $%d)SELECT total_count, id, name, email, phone_no, status,type,archived_at,archive_reason,deleted_by,asset_quantity FROM cte_employee", args+1, args+2)
	SQL += pageStr
	values = append(values, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)

	var getEmployee = make([]models.GetEmployee, 0)
	err := database.AssetManagement.Select(&getEmployee, SQL, values...)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("GetEmployee: cannot get employee list.")
		return totalGetEmployee, err
	}

	totalGetEmployee.GetEmployee = getEmployee
	if len(getEmployee) == 0 {
		logrus.WithError(err).Error("GetEmployee: got empty array.")
		return totalGetEmployee, err
	}

	totalGetEmployee.TotalCount = getEmployee[0].TotalCount
	return totalGetEmployee, nil
}

func UpdateEmployee(user *models.EmployeeDetails) error {
	SQL := `UPDATE employee
            SET name       = $1,
                email      = $2,
                phone_no   = $3,
                updated_at = NOW(),
                status     = $5,
                type       = $6
            WHERE id = $4
              AND archived_at IS NULL`
	_, err := database.AssetManagement.Exec(SQL, user.Name, user.Email, user.PhoneNo, user.ID, user.Status, user.Type)
	if err != nil {
		logrus.WithError(err).Error("UpdateEmployee: cannot update employee details.")
		return err
	}
	return nil
}

func DeleteEmployee(employeeID, userID string, employeeBody models.Employee) error {
	SQL := `UPDATE employee
            SET    archived_at = now(),
                   archive_reason = $2,
                   deleted_by = $3,
                   status = $4
            WHERE  id = $1
            AND    archived_at IS NULL 
            `
	_, err := database.AssetManagement.Exec(SQL, employeeID, employeeBody.ArchiveReason, userID, utils.Deleted)
	if err != nil {
		logrus.WithError(err).Error("DeleteEmployee: cannot delete employee.")
		return err
	}
	return nil
}

func CreateEmployeeAssetRelation(employeeAssetRelation models.EmployeeAssetRelation, assignedBy string, tx *sqlx.Tx) error {
	SQL := `INSERT INTO employee_asset_relation(employee_id, asset_id, assigned_by, assigned_date)
            VALUES ($1, $2, $3, $4)`

	_, err := tx.Exec(SQL, employeeAssetRelation.EmployeeID, employeeAssetRelation.AssetID, assignedBy, employeeAssetRelation.AssignedDate)
	if err != nil {
		logrus.WithError(err).Error("CreateEmployeeAssetRelation: cannot create employee asset relation.")
		return err
	}
	return nil
}

func GetAssetHistory(employeeID string) ([]models.AssetHistory, error) {
	SQL := `SELECT  a.id, 
       				brand, 
       				model, 
       				serial_no, 
       				asset_type, 
       				assigned_date, 
       				retrieved_date,
       				coalesce(retrieval_reason, '') as retrieval_reason
			FROM employee_asset_relation ear
			JOIN assets a ON ear.asset_id = a.id
			WHERE a.archived_at IS NULL
			AND (
			           NULLIF(LENGTH($1),0) IS NULL 
			           OR ear.employee_id::text = $1
			)
`
	assetHistory := make([]models.AssetHistory, 0)
	err := database.AssetManagement.Select(&assetHistory, SQL, employeeID)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("GetAssetHistory: cannot get asset history.")
		return nil, err
	}
	return assetHistory, nil
}
