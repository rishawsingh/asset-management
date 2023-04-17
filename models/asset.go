package models

import (
	"time"

	"github.com/volatiletech/null"
)

type CreateAsset struct {
	Brand              string      `json:"brand" db:"brand" validate:"required"`
	Model              string      `json:"model" db:"model"`
	SerialNo           string      `json:"serialNo" db:"serial_no"`
	AssetType          AssetType   `json:"AssetType" db:"asset_type" validate:"required"`
	PurchasedDate      time.Time   `json:"purchasedDate" db:"purchased_date" validate:"required"`
	WarrantyStartDate  time.Time   `json:"warrantyStartDate" db:"warranty_start_date" validate:"required"`
	WarrantyExpiryDate time.Time   `json:"warrantyExpiryDate" db:"warranty_expiry_date" validate:"required"`
	Series             string      `json:"series" db:"series"`
	Processor          string      `json:"processor" db:"processor"`
	RAM                string      `json:"ram" db:"ram"`
	OperatingSystem    string      `json:"operatingSystem" db:"operating_system"`
	Charger            bool        `json:"charger" db:"charger"`
	ScreenResolution   string      `json:"screenResolution" db:"screen_resolution"`
	Storage            string      `json:"storage" db:"storage"`
	OsType             string      `json:"osType" db:"os_type"`
	Imei1              string      `json:"imei1" db:"imei_1"`
	Imei2              string      `json:"imei2" db:"imei_2"`
	SimNo              string      `json:"simNo" db:"sim_no"`
	PhoneNo            string      `json:"phoneNo" db:"phone_no"`
	OwnedBy            string      `json:"ownedBy" db:"owned_by"`
	ClientName         string      `json:"clientName" db:"client_name"`
	Status             string      `json:"status" db:"status"`
	ArchivedAt         null.Time   `json:"archivedAt" db:"archived_at"`
	ArchiveReason      null.String `json:"archiveReason" db:"archive_reason"`
	DeletedBy          null.String `json:"deletedBy" db:"deleted_by"`
	AssetHistory       []EmployeeHistory
}

type TotalGetAsset struct {
	GetAsset   []GetAsset
	TotalCount int `json:"totalCount" db:"total_count"`
}

type GetAsset struct {
	TotalCount         int         `json:"-" db:"total_count"`
	ID                 string      `json:"id" db:"id"`
	Brand              string      `json:"brand" db:"brand"`
	Model              string      `json:"model" db:"model"`
	SerialNo           string      `json:"serialNo" db:"serial_no"`
	AssetType          AssetType   `json:"AssetType" db:"asset_type"`
	PurchasedDate      time.Time   `json:"purchasedDate" db:"purchased_date"`
	WarrantyStartDate  time.Time   `json:"warrantyStartDate" db:"warranty_start_date"`
	WarrantyExpiryDate time.Time   `json:"warrantyExpiryDate" db:"warranty_expiry_date"`
	AssignedToID       null.String `json:"assignedToID" db:"assigned_to_id"`
	AssignedTo         null.String `json:"assignedTo" db:"name"`
	Status             string      `json:"status" db:"status"`
}

type UpdateAssetSpecification struct {
	Brand              string    `json:"brand" db:"brand" validate:"required"`
	Model              string    `json:"model" db:"model"`
	SerialNo           string    `json:"serialNo" db:"serial_no"`
	PurchasedDate      time.Time `json:"purchasedDate" db:"purchased_date" validate:"required"`
	WarrantyStartDate  time.Time `json:"warrantyStartDate" db:"warranty_start_date"`
	WarrantyExpiryDate time.Time `json:"warrantyExpiryDate" db:"warranty_expiry_date"`
	Series             string    `json:"series" db:"series"`
	Processor          string    `json:"processor" db:"processor"`
	RAM                string    `json:"ram" db:"ram"`
	OperatingSystem    string    `json:"operatingSystem" db:"operating_system"`
	Charger            bool      `json:"charger" db:"charger"`
	ScreenResolution   string    `json:"screenResolution" db:"screen_resolution"`
	Storage            string    `json:"storage" db:"storage"`
	ID                 string    `json:"id" db:"id" validate:"required"`
	AssetType          AssetType `json:"AssetType" db:"asset_type" validate:"required"`
	OsType             string    `json:"osType" db:"os_type"`
	Imei1              string    `json:"imei1" db:"imei_1"`
	Imei2              string    `json:"imei2" db:"imei_2"`
	SimNo              string    `json:"simNo" db:"sim_no"`
	PhoneNo            string    `json:"phoneNo" db:"phone_no"`
}

type ReassignAsset struct {
	AssetID         string    `json:"assetId" db:"asset_id" validate:"required"`
	EmployeeID      string    `json:"employeeId" db:"employee_id" validate:"required"`
	RetrievedDate   time.Time `json:"retrievedDate" db:"retrieved_date" validate:"required"`
	RetrievalReason string    `json:"retrievalReason" db:"retrieval_reason" validate:"required"`
	AssignedDate    string    `json:"assignedDate" db:"assigned_date" validate:"required"`
}

type AssignAssetDetails struct {
	ID       string `json:"Id" db:"id"`
	Imei1    string `json:"imei1" db:"imei_1"`
	SerialNo string `json:"serialNo" db:"serial_no"`
	Model    string `json:"model" db:"model"`
	Brand    string `json:"brand" db:"brand"`
	SimNo    string `json:"simNo" db:"sim_no"`
}

type EmployeeHistory struct {
	ID              string    `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Email           string    `json:"email" db:"email"`
	PhoneNo         string    `json:"phoneNo" db:"phone_no"`
	AssetType       string    `json:"assetType" db:"asset_type"`
	AssignedDate    string    `json:"assignedDate" db:"assigned_date"`
	AssignedBy      string    `json:"assignedBy" db:"assigned_by"`
	RetrievedDate   null.Time `json:"retrievedDate" db:"retrieved_date"`
	RetrievalReason string    `json:"retrievalReason" db:"retrieval_reason"`
}

type WarrantyDetails struct {
	AssetID            string    `json:"assetId"`
	WarrantyStartDate  time.Time `json:"warrantyStartDate" db:"warranty_start_date"`
	WarrantyExpiryDate time.Time `json:"warrantyExpiryDate" db:"warranty_expiry_date"`
}

type AssetRetrievalDetails struct {
	RetrievedDate   time.Time `json:"retrievedDate" db:"retrieved_date"`
	RetrievalReason string    `json:"retrievalReason" db:"retrieval_reason"`
	EmployeeID      string    `json:"employeeId" db:"employee_id"`
	AssetID         string    `json:"assetId" db:"asset_id"`
}

type Asset struct {
	ID           string    `json:"id" db:"asset_id" validate:"required"`
	AssetType    AssetType `json:"assetType" db:"asset_type" validate:"required"`
	DeleteReason string    `json:"deleteReason" db:"archive_reason"`
}
