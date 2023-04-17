package models

import (
	"time"

	"github.com/lib/pq"

	"github.com/dgrijalva/jwt-go"
	"github.com/volatiletech/null"
)

type LoginDetails struct {
	NoOfTime  int       `json:"noOfTime" db:"no_of_time"`
	StartTime time.Time `json:"startTime" db:"start_time"`
}

type AccessedByDetails struct {
	ID                  string    `json:"Id" db:"id"`
	Name                string    `json:"name" db:"name"`
	Email               string    `json:"email" db:"email"`
	AuthenticationTimes int       `json:"authenticationTimes" db:"authentication_times"`
	LastLoginTime       null.Time `json:"lastLoginTime" db:"start_time"`
	Status              string    `json:"status" db:"status"`
}

type FiltersCheck struct {
	Pagination    bool
	IsExpired     bool
	IsSearched    bool
	SearchedName  string
	EmployeeID    string
	Limit         int
	Page          int
	AssetTypes    pq.StringArray
	Available     bool
	NotAnEmployee bool
	Deleted       bool
	Assigned      bool
	Warranty      int
}

type AssetType string

const (
	Laptop   AssetType = "laptop"
	Pendrive AssetType = "pen drive"
	Harddisk AssetType = "hard disk"
	Mouse    AssetType = "mouse"
	Mobile   AssetType = "mobile"
	Sim      AssetType = "sim"
)

type RegisterUser struct {
	Name     string `json:"name" db:"name" validate:"required"`
	Email    string `json:"email" db:"email" validate:"required,email"`
	PhoneNo  string `json:"phoneNo" db:"phone_no" validate:"required,min=10,max=10,numeric"`
	Password string `json:"password" db:"password" validate:"required,min=6"`
}

type UsersLoginDetails struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type StatusDetails struct {
	Type                string `json:"type" db:"type"`
	AuthenticationTimes int    `json:"authenticationTimes" db:"authentication_times"`
	Status              string `json:"status" db:"status"`
}

type UserCredentials struct {
	ID       string `json:"id" db:"id"`
	Password string `json:"password" db:"password"`
}

type Claims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

type ContextValues struct {
	ID string `json:"id"`
}

type EmployeeID struct {
	ID string `json:"id" db:"employee_id" validate:"required"`
}

type GetAssetByID struct {
	Brand         string    `json:"brand" db:"brand"`
	Model         string    `json:"model" db:"model"`
	SerialNo      string    `json:"serialNo" db:"serial_no"`
	PurchasedDate time.Time `json:"purchasedDate" db:"purchased_date"`
}

type UpdateAsset struct {
	ID            string    `json:"id" db:"id" validate:"required"`
	Brand         string    `json:"brand" db:"brand" validate:"required"`
	Model         string    `json:"model" db:"model" validate:"required"`
	SerialNo      string    `json:"serialNo" db:"serial_no" validate:"required"`
	PurchasedDate time.Time `json:"purchasedDate" db:"purchased_date" validate:"required"`
}

type GetAssetQuantity struct {
	TotalAssets       int `json:"totalAssets" db:"total_assets"`
	DistributedAssets int `json:"distributedAssets" db:"distributed_assets"`
	AvailableAssets   int `json:"availableAssets" db:"available_assets"`
	LaptopQuantity    int `json:"laptopQuantity" db:"laptop_quantity"`
	MouseQuantity     int `json:"mouseQuantity" db:"mouse_quantity"`
	PenDriveQuantity  int `json:"penDriveQuantity" db:"pen_drive_quantity"`
	HardDiskQuantity  int `json:"hardDiskQuantity" db:"hard_disk_quantity"`
	MobileQuantity    int `json:"mobileQuantity" db:"mobile_quantity"`
	SimQuantity       int `json:"simQuantity" db:"sim_quantity"`
}

type UserDetails struct {
	Name    string `json:"name" db:"name"`
	Email   string `json:"email" db:"email"`
	PhoneNo string `json:"phoneNo" db:"phone_no"`
	Image   string `json:"image" db:"image"`
}
