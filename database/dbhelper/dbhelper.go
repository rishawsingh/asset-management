package dbhelper

import (
	"InternalAssetManagement/database"
	"InternalAssetManagement/models"
	"database/sql"

	"github.com/sirupsen/logrus"
)

func AddProfileImage(userID, url string) error {
	SQL := `UPDATE users
            SET    image = $1
            WHERE  id = $2
            AND    archived_at IS NULL 
            `
	_, err := database.AssetManagement.Exec(SQL, url, userID)
	if err != nil {
		logrus.WithError(err).Error("AddProfileImage: cannot add image.")
		return err
	}
	return nil
}

func AlterStatusDetails(userType, status string, authenticationTimes int, userID string) error {
	SQL := `UPDATE users
            SET    authentication_times = $1,
                   status = $2, 
                   type = $3
            WHERE  id = $4
            `
	_, err := database.AssetManagement.Exec(SQL, authenticationTimes+1, status, userType, userID)
	if err != nil {
		logrus.WithError(err).Error("AlterStatusDetails: cannot alter authentication times.")
		return err
	}
	return nil
}

func GetStatusDetails(email string) (models.StatusDetails, error) {
	SQL := `SELECT  status,
       				type,
       				authentication_times
            FROM    users
            WHERE   email = $1
           `
	var statusDetails models.StatusDetails

	err := database.AssetManagement.Get(&statusDetails, SQL, email)
	if err != nil {
		logrus.WithError(err).Error("GetStatusDetails: cannot get user status details.")
		return statusDetails, err
	}
	return statusDetails, nil
}

func AccessedByDetails(userType string, filterCheck *models.FiltersCheck) ([]models.AccessedByDetails, error) {
	SQL := `SELECT  users.id as id,
       				name,
       				email,
       				authentication_times,
       				status,
       				max(start_time) as start_time
			FROM   users LEFT JOIN sessions s on users.id = s.user_id
			WHERE
			CASE WHEN $1 = 'authorized' THEN type = 'authorized' ELSE type != 'authorized' END
				AND  ($3 OR (name ilike '%%' || $2 || '%%'))
				GROUP BY ( users.id,name,email,authentication_times,status)
			`

	accessedByDetails := make([]models.AccessedByDetails, 0)
	err := database.AssetManagement.Select(&accessedByDetails, SQL, userType, filterCheck.SearchedName, !filterCheck.IsSearched)
	if err != nil {
		logrus.WithError(err).Error("AccessedByDetails: cannot accessed by details.")
		return accessedByDetails, err
	}
	return accessedByDetails, nil
}

func Logout(userID string) error {
	SQL := `UPDATE sessions
            SET    end_time=now()
            WHERE  user_id=$1`

	_, err := database.AssetManagement.Exec(SQL, userID)
	if err != nil {
		logrus.WithError(err).Error("Logout: cannot do logout.")
		return err
	}
	return nil
}

func IsUserExist(email, phoneNo string) (bool, error) {
	SQL := `SELECT id FROM users 
            where email = $1 
            AND phone_no = $2 
            AND archived_at IS NULL`
	var id string
	err := database.AssetManagement.Get(&id, SQL, email, phoneNo)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("IsUserExist: cannot get if user exist or not.")
		return false, err
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return true, nil
}

func CreateUser(name, email, password, phoneNo string) error {
	SQL := `INSERT INTO users (name, email, password, phone_no) 
            VALUES ($1,$2,$3,$4)`
	_, err := database.AssetManagement.Exec(SQL, name, email, password, phoneNo)
	if err != nil {
		logrus.WithError(err).Error("CreateUser: cannot create user.")
		return err
	}
	return nil
}

func FetchPasswordAndID(email string) (models.UserCredentials, error) {
	SQL := `SELECT  users.id,
       				password
            FROM   users
            WHERE  email=$1 
            `

	var userCredentials models.UserCredentials

	err := database.AssetManagement.Get(&userCredentials, SQL, email)
	if err != nil {
		logrus.WithError(err).Error("FetchPasswordAndID: Not able to fetch password, ID.")
		return userCredentials, err
	}
	return userCredentials, nil
}

func CreateSession(claims *models.Claims) error {
	SQL := `INSERT INTO sessions(user_id)
            VALUES   ($1)`
	_, err := database.AssetManagement.Exec(SQL, claims.ID)
	if err != nil {
		logrus.WithError(err).Error("CreateSession: cannot create user session.")
		return err
	}
	return nil
}

func CheckSession(userID string) (string, error) {
	SQL := `SELECT id
           FROM    sessions
           WHERE   sessions.end_time IS NULL
           AND     user_id=$1
           ORDER BY start_time DESC
           LIMIT 1`
	var sessionID string

	err := database.AssetManagement.Get(&sessionID, SQL, userID)
	if err != nil {
		logrus.WithError(err).Error("CheckSession: session expired.")
		return sessionID, err
	}
	return sessionID, nil
}

func GetUserDetails(id string) (*models.UserDetails, error) {
	SQL := `SELECT name, email, phone_no, coalesce(image, '') as image
			FROM users
			WHERE id = $1
			  AND archived_at IS NULL`
	var user models.UserDetails
	err := database.AssetManagement.Get(&user, SQL, id)
	if err != nil {
		logrus.WithError(err).Error("GetUserDetails: cannot get user details.")
		return nil, err
	}
	return &user, nil
}

func GetTotalAssetQuantities(assetQuantities models.GetAssetQuantity) (models.GetAssetQuantity, error) {
	SQL := `WITH cte_total AS(
    							SELECT count(id) AS total_assets    
    							FROM   assets
    							WHERE archived_at IS NULL
							 ),  
    	  cte_distributed AS(
    							SELECT count(asset_id) AS distributed_assets
    							FROM  employee_asset_relation
    							WHERE retrieved_date IS NULL
							)
		  SELECT ct.total_assets AS total_assets,
       			 cd.distributed_assets AS distributed_assets,
       			 ct.total_assets - cd.distributed_assets AS available_assets
		  FROM cte_total ct, cte_distributed cd
		  `
	err := database.AssetManagement.Get(&assetQuantities, SQL)
	if err != nil {
		logrus.WithError(err).Error("GetTotalAssetQuantities: cannot get total asset quantities.")
		return assetQuantities, err
	}
	return assetQuantities, nil
}

func GetAssetQuantities(dashBoardFilter string) (models.GetAssetQuantity, error) {
	SQL := `SELECT count(*) filter ( where asset_type = 'laptop' )    AS laptop_quantity,
				   count(*) filter ( where asset_type = 'mouse' )     AS mouse_quantity,
				   count(*) filter ( where asset_type = 'pen drive' ) AS pen_drive_quantity,
				   count(*) filter ( where asset_type = 'hard disk' ) AS hard_disk_quantity,
				   count(*) filter ( where asset_type = 'mobile' ) AS mobile_quantity,
				   count(*) filter ( where asset_type = 'sim' ) AS sim_quantity
			FROM assets a
			`

	switch dashBoardFilter {
	case "distributed":
		SQL += "JOIN employee_asset_relation ear on a.id = ear.asset_id WHERE a.archived_at IS NULL AND ear.retrieved_date IS NULL"
	case "available":
		SQL += "WHERE a.archived_at IS NULL AND a.is_available = true"
	case "total":
		SQL += "WHERE a.archived_at IS NULL"
	}
	var assetQuantity models.GetAssetQuantity
	err := database.AssetManagement.Get(&assetQuantity, SQL)
	if err != nil {
		logrus.WithError(err).Error("GetAssetQuantities: cannot get asset quantities.")
		return assetQuantity, err
	}
	return assetQuantity, nil
}

func UpdateUser(user models.RegisterUser, password, id string) error {
	SQL := `UPDATE users
            SET name       = $1,
                email      = $2,
                phone_no   = $3,
                password   = $4,
                updated_at = NOW()
            WHERE id = $5
              AND archived_at IS NULL`
	_, err := database.AssetManagement.Exec(SQL, user.Name, user.Email, user.PhoneNo, password, id)
	if err != nil {
		logrus.WithError(err).Error("UpdateUser: cannot update user details.")
		return err
	}
	return nil
}

func UpdateAccessedBy(userID, userType string) error {
	SQL := `UPDATE users
            SET   type = $1
            WHERE id = $2
            AND   archived_at IS NULL 
            `
	_, err := database.AssetManagement.Exec(SQL, userType, userID)
	if err != nil {
		logrus.WithError(err).Error("UpdateAccessedBy: cannot update AccessedBy.")
		return err
	}
	return nil
}
