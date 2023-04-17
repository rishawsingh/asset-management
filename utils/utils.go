package utils

import (
	"InternalAssetManagement/models"
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	cloud "cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
)

type Key string

const (
	Whitelisted   = "white_listed"
	Blocklisted   = "black_listed"
	Warned        = "warned"
	Authorized    = "authorized"
	UnAuthorized  = "unauthorized"
	Blocked       = "blocked"
	Laptop        = "laptop"
	Pendrive      = "pen drive"
	Harddisk      = "hard disk"
	Mouse         = "mouse"
	Mobile        = "mobile"
	Sim           = "sim"
	RemoteState   = "remote_state"
	Available     = "available"
	Assigned      = "assigned"
	Deleted       = "deleted"
	Active        = "active"
	NotAnEmployee = "not_an_employee"
)

const UserContextKey Key = "userID"

var generator *shortid.Shortid

const generatorSeed = 1000

const DefaultLimit = 10

type FieldError struct {
	Err validator.ValidationErrors
}

type ResponseMsg struct {
	Msg string `json:"msg"`
}

func (q FieldError) GetSingleError() string {
	errorString := ""
	for _, e := range q.Err {
		errorString = "Invalid " + e.Field()
	}
	return errorString
}

type clientError struct {
	ID            string `json:"id"`
	MessageToUser string `json:"messageToUser"`
	DeveloperInfo string `json:"developerInfo"`
	Err           string `json:"error"`
	StatusCode    int    `json:"statusCode"`
	IsClientError bool   `json:"isClientError"`
}

func init() {
	n, err := rand.Int(rand.Reader, big.NewInt(generatorSeed))
	if err != nil {
		logrus.Panicf("failed to initialize utilities with random seed, %+v", err)
		return
	}

	g, err := shortid.New(1, shortid.DefaultABC, n.Uint64())

	if err != nil {
		logrus.Panicf("Failed to initialize utils package with error: %+v", err)
	}

	generator = g
}

// ParseBody parses the values from io reader to a given interface
func ParseBody(body io.Reader, out interface{}) error {
	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		return err
	}

	return nil
}

// EncodeJSONBody writes the JSON body to response writer
func EncodeJSONBody(resp http.ResponseWriter, data interface{}) error {
	return json.NewEncoder(resp).Encode(data)
}

// RespondJSON sends the interface as a JSON
func RespondJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if body != nil {
		if err := EncodeJSONBody(w, body); err != nil {
			logrus.Errorf("Failed to respond JSON with error: %+v", err)
		}
	}
}

// newClientError creates structured client error response message
func newClientError(err error, statusCode int, messageToUser string, additionalInfoForDevs ...string) *clientError {
	additionalInfoJoined := strings.Join(additionalInfoForDevs, "\n")
	if additionalInfoJoined == "" {
		additionalInfoJoined = messageToUser
	}

	errorID, _ := generator.Generate()
	var errString string
	if err != nil {
		errString = err.Error()
	}
	return &clientError{
		ID:            errorID,
		MessageToUser: messageToUser,
		DeveloperInfo: additionalInfoJoined,
		Err:           errString,
		StatusCode:    statusCode,
		IsClientError: true,
	}
}

// RespondError sends an error message to the API caller and logs the error
func RespondError(w http.ResponseWriter, statusCode int, err error, messageToUser string, additionalInfoForDevs ...string) {
	logrus.Errorf("status: %d, message: %s, err: %+v ", statusCode, messageToUser, err)
	clientError := newClientError(err, statusCode, messageToUser, additionalInfoForDevs...)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(clientError); err != nil {
		logrus.Errorf("Failed to send error to caller with error: %+v", err)
	}
}

// HashString generates SHA256 for a given string
func HashString(toHash string) string {
	sha := sha512.New()
	sha.Write([]byte(toHash))
	return hex.EncodeToString(sha.Sum(nil))
}

// HashPassword returns the bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if the provided password is correct or not
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func CheckPhoneNo(phoneNo string) error {
	_, err := strconv.Atoi(phoneNo)
	if err != nil {
		return err
	}
	return nil
}

func UserContext(r *http.Request) (string, error) {
	user := r.Context().Value(UserContextKey)
	userID, ok := user.(string)
	if !ok {
		return "", errors.New("unable to convert userID")
	}
	return userID, nil
}

// CheckValidation returns the current validation status
func CheckValidation(i interface{}) validator.ValidationErrors {
	v := validator.New()
	err := v.Struct(i)
	if err == nil {
		return nil
	}
	return err.(validator.ValidationErrors)
}

// TrimAll removes a given rune form given string
func TrimAll(str string, remove rune) string {
	return strings.Map(func(r rune) rune {
		if r == remove {
			return -1
		}
		return r
	}, str)
}

// TrimStringAfter trims anything after given delimiter
func TrimStringAfter(s, delim string) string {
	if idx := strings.Index(s, delim); idx != -1 {
		return s[:idx]
	}
	return s
}

type FirebaseApp struct {
	Ctx     context.Context
	Client  *firestore.Client
	Storage *cloud.Client
}

func UploadImage(request *http.Request) (string, error) {
	client := FirebaseApp{}
	var err error
	client.Ctx = context.Background()
	credentialsFile := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
	// fmt.Println(credentialsFile)
	app, err := firebase.NewApp(client.Ctx, nil, credentialsFile)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	client.Client, err = app.Firestore(client.Ctx)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	client.Storage, err = cloud.NewClient(client.Ctx, credentialsFile)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	file, fileHeader, err := request.FormFile("image")
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	var multiPartMinValue, multiPartMaxValue int64
	multiPartMinValue = 10
	multiPartMaxValue = 20
	err = request.ParseMultipartForm(multiPartMinValue << multiPartMaxValue)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	defer func(file multipart.File) {
		fileErr := file.Close()
		if fileErr != nil {
			return
		}
	}(file)
	imagePath := "images/" + fileHeader.Filename
	bucket := "storex-cd365.appspot.com"
	bucketStorage := client.Storage.Bucket(bucket).Object(imagePath).NewWriter(client.Ctx)

	_, err = io.Copy(bucketStorage, file)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	if err1 := bucketStorage.Close(); err1 != nil {
		logrus.Error(err1)
		return "", err
	}

	hours := 100

	signedURL := &cloud.SignedURLOptions{
		Scheme:  cloud.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(time.Duration(hours) * time.Hour),
	}

	url, err := client.Storage.Bucket(bucket).SignedURL(imagePath, signedURL)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	return url, nil
}

func ParamStrToBool(paramStrValue string) (bool, error) {
	var paramValue bool
	var err error
	if paramStrValue == "true" || paramStrValue == "TRUE" || paramStrValue == "True" {
		paramValue, err = strconv.ParseBool(paramStrValue)
		if err != nil {
			logrus.Printf("unable to convert string to bool:%v", err)
			return paramValue, err
		}
	}
	return paramValue, nil
}

func Filters(r *http.Request) (models.FiltersCheck, error) {
	filtersCheck := models.FiltersCheck{}
	isSearched := false
	isExpired := false
	searchedName := r.URL.Query().Get("name")
	if searchedName != "" {
		isSearched = true
	}
	var warrantyAsset int
	var limit int
	var err error
	var page int
	strLimit := r.URL.Query().Get("limit")
	if strLimit == "" {
		limit = DefaultLimit
	} else {
		limit, err = strconv.Atoi(strLimit)
		if err != nil {
			logrus.Printf("Limit: cannot get limit:%v", err)
			return filtersCheck, err
		}
	}
	var pagination bool
	strPagination := r.URL.Query().Get("pagination")
	if strPagination == "" {
		pagination = true
	} else {
		pagination, err = strconv.ParseBool(strPagination)
		if err != nil {
			logrus.Printf("unable to convert string to bool:%v", err)
			return filtersCheck, err
		}
	}

	var availableAssets bool
	strAvailable := r.URL.Query().Get("available")
	availableAssets, err = ParamStrToBool(strAvailable)
	if err != nil {
		logrus.Printf("unable to convert string to bool:%v", err)
		return filtersCheck, err
	}

	var assignedAssets bool
	strAssignedAsset := r.URL.Query().Get("assigned")
	assignedAssets, err = ParamStrToBool(strAssignedAsset)
	if err != nil {
		logrus.Printf("unable to convert string to bool:%v", err)
		return filtersCheck, err
	}

	var deleted bool

	strDeleted := r.URL.Query().Get("deleted")
	deleted, err = ParamStrToBool(strDeleted)
	if err != nil {
		logrus.Printf("unable to convert string to bool:%v", err)
		return filtersCheck, err
	}

	var notAnEmployee bool

	strNotAnEmployee := r.URL.Query().Get("notAnEmployee")
	notAnEmployee, err = ParamStrToBool(strNotAnEmployee)
	if err != nil {
		logrus.Printf("unable to convert string to bool:%v", err)
		return filtersCheck, err
	}

	var assetTypes []string

	assetType := r.URL.Query().Get("assetType")
	if assetType != "" {
		assetTypeFilter := strings.Split(assetType, ",")
		filterTypes := make([]string, 0)
		for i := range assetTypeFilter {
			filterTypes = append(filterTypes, assetTypeFilter[i])
		}
		assetTypes = filterTypes
	}

	warranty := r.URL.Query().Get("warranty")
	if warranty == "" {
		warrantyAsset = 0
	} else {
		warrantyAsset, err = strconv.Atoi(warranty)
		if err != nil {
			logrus.Printf("Warranty: cannot convert string to int")
			return filtersCheck, err
		}
		if warrantyAsset == 0 {
			isExpired = true
		}
	}

	strPage := r.URL.Query().Get("page")
	if strPage == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(strPage)
		if err != nil {
			logrus.Printf("Page: cannot get page:%v", err)
			return filtersCheck, err
		}
	}

	filtersCheck = models.FiltersCheck{
		AssetTypes:    assetTypes,
		IsSearched:    isSearched,
		SearchedName:  searchedName,
		Page:          page,
		Limit:         limit,
		Available:     availableAssets,
		Deleted:       deleted,
		Assigned:      assignedAssets,
		NotAnEmployee: notAnEmployee,
		Warranty:      warrantyAsset,
		IsExpired:     isExpired,
		Pagination:    pagination}
	return filtersCheck, nil
}
