package models

import (
	"time"

	"github.com/volatiletech/null"
)

type EmployeeDetails struct {
	ID      string `json:"id" db:"id"`
	Name    string `json:"name" db:"name" validate:"required"`
	Type    string `json:"type" db:"type" validate:"required"`
	Email   string `json:"email" db:"email" validate:"required,email"`
	PhoneNo string `json:"phoneNo" db:"phone_no" validate:"required,min=10,max=10,numeric"`
	Status  string `json:"status" db:"status"`
}

type TotalGetEmployee struct {
	GetEmployee []GetEmployee
	TotalCount  int `json:"totalCount" db:"total_count"`
}

type GetEmployee struct {
	TotalCount    int            `json:"-" db:"total_count"`
	ID            string         `json:"id" db:"id"`
	Name          string         `json:"name" db:"name"`
	Email         string         `json:"email" db:"email"`
	PhoneNo       string         `json:"phoneNo" db:"phone_no"`
	Status        string         `json:"status" db:"status"`
	Type          string         `json:"type" db:"type"`
	ArchivedAt    null.Time      `json:"archivedAt" db:"archived_at"`
	ArchiveReason null.String    `json:"archiveReason" db:"archive_reason"`
	DeletedBy     null.String    `json:"deletedBy" db:"deleted_by"`
	AssetQuantity int            `json:"assetQuantity" db:"asset_quantity"`
	AssetHistory  []AssetHistory `json:"assetHistory"`
}

type EmployeeAssetRelation struct {
	EmployeeID   string    `json:"employeeID" db:"employee_id"`
	AssetID      string    `json:"assetId" db:"asset_id"`
	AssignedDate time.Time `json:"assignedDate" db:"assigned_date"`
}

type AssetHistory struct {
	ID              string     `json:"id" db:"id"`
	Brand           string     `json:"brand" db:"brand"`
	Model           string     `json:"model" db:"model"`
	SerialNo        string     `json:"serialNo" db:"serial_no"`
	AssetType       AssetType  `json:"AssetType" db:"asset_type"`
	AssignedDate    time.Time  `json:"assignedDate" db:"assigned_date"`
	RetrievedDate   *time.Time `json:"retrievedDate" db:"retrieved_date"`
	RetrievalReason string     `json:"retrievalReason" db:"retrieval_reason"`
}

type GetEmployeeList struct {
	ID            string  `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Email         string  `json:"email" db:"email"`
	PhoneNo       string  `json:"phoneNo" db:"phone_no"`
	AssetQuantity int     `json:"assetQuantity" db:"asset_quantity"`
	AssetHistory  []uint8 `json:"assetHistory" db:"asset_history"`
}

type AssetHistoryDetails struct {
	AssetID       string `json:"assetId" db:"asset_id"`
	Brand         string `json:"brand" db:"brand"`
	Model         string `json:"model" db:"model"`
	SerialNo      string `json:"serialNo" dbb:"serial_no"`
	AssetType     string `json:"assetType" db:"asset_type"`
	AssignedDate  string `json:"assignedDate" db:"assigned_date"`
	RetrievedDate string `json:"retrievedDate" db:"retrieved_date"`
}

type Employee struct {
	ArchiveReason string `json:"archiveReason" db:"archive_reason"`
}
