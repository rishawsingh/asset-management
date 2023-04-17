package handler

import (
	"InternalAssetManagement/database/dbhelper"
	"InternalAssetManagement/models"
	"InternalAssetManagement/utils"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func AddProfileImage(w http.ResponseWriter, r *http.Request) {
	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "Cannot get user id.")
		return
	}

	url, err := utils.UploadImage(r)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "UploadImage: cannot upload image url.")
		return
	}
	fmt.Println(url)
	err = dbhelper.AddProfileImage(userID, url)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "AddProfileImage: cannot add profile image.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Msg string `json:"msg"`
		URL string `json:"url"`
	}{
		Msg: "Added image successfully.",
		URL: url,
	})
}

func AccessedByDetails(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query().Get("userType")
	filterCheck, err := utils.Filters(r)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "AccessedByDetails: cannot get filters properly: ")
		return
	}

	accessedByDetails, err := dbhelper.AccessedByDetails(userType, &filterCheck)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "AccessedByDetails: cannot get Accessed By Details.")
		return
	}
	if accessedByDetails == nil {
		utils.RespondJSON(w, http.StatusOK, []models.AccessedByDetails{})
		return
	}
	utils.RespondJSON(w, http.StatusOK, accessedByDetails)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "Cannot get user id.")
		return
	}

	err := dbhelper.Logout(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Logout: unable to logout.")
		return
	}
}

var JwtKey = []byte("secret_key")

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	body := models.RegisterUser{}
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "Failed to parse request body.")
		return
	}

	validationErr := validate.Struct(body)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error")
		return
	}

	exists, existsErr := dbhelper.IsUserExist(body.Email, body.PhoneNo)
	if existsErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existsErr, "failed to check users' existence.")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusBadRequest, nil, "user already exists.")
		return
	}

	hashedPassword, hashErr := utils.HashPassword(body.Password)
	if hashErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, hashErr, "failed to secure password.")
		return
	}

	err := dbhelper.CreateUser(body.Name, body.Email, hashedPassword, body.PhoneNo)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "failed to create user.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "User registered.",
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var userDetails models.UsersLoginDetails
	decoderErr := utils.ParseBody(r.Body, &userDetails)
	if decoderErr != nil {
		utils.RespondError(w, http.StatusBadRequest, decoderErr, "LoginUser: decoder error.")
		return
	}

	companyEmail := strings.Split(userDetails.Email, "@")[1]

	if companyEmail != "remotestate.com" {
		utils.RespondJSON(w, http.StatusBadRequest, utils.ResponseMsg{
			Msg: "non-authorized email.",
		})
		utils.RespondError(w, http.StatusBadRequest, errors.New("non-authorized email"), "LoginUser: non-authorized email")
		return
	}

	userCredentials, fetchErr := dbhelper.FetchPasswordAndID(userDetails.Email)
	if fetchErr != nil {
		if fetchErr == sql.ErrNoRows {
			utils.RespondJSON(w, http.StatusBadRequest, utils.ResponseMsg{
				Msg: "wrong email.",
			})

			logrus.Printf("Login:FetchPasswordAndId: wrong details:%v", fetchErr)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if PasswordErr := bcrypt.CompareHashAndPassword([]byte(userCredentials.Password), []byte(userDetails.Password)); PasswordErr != nil {
		_, err := w.Write([]byte("ERROR: Wrong password"))
		if err != nil {
			return
		}
		utils.RespondError(w, http.StatusUnauthorized, PasswordErr, "Login: Password misMatch")
		return
	}

	statusDetails, err := dbhelper.GetStatusDetails(userDetails.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "LoginUser: unable to fetch user status.")
		return
	}
	switch {
	case statusDetails.Type == utils.UnAuthorized && statusDetails.AuthenticationTimes == 0:

		// Send Email
		from := mail.NewEmail("me", "tushar.tushid@remotestate.com")
		to := mail.NewEmail("user", userDetails.Email)
		subject := "Email received through twilio sendgrid"
		plainTextContent := "WARNING: unauthorized email"
		htmlContent := "<strong> and easy to do with go!"
		message1 := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
		client1 := sendgrid.NewSendClient(os.Getenv("sendgrid_api_key"))
		response1, RErr := client1.Send(message1)
		if RErr != nil {
			logrus.Printf("SendFriendRequest: cannot send mail to user:%v", RErr)
			return
		}

		fmt.Println(response1.StatusCode)
		fmt.Println(response1.Body)
		fmt.Println(response1.Headers)

		// send email over
		utils.RespondJSON(w, http.StatusBadRequest, utils.ResponseMsg{
			Msg: "unauthorized email.",
		})

		err = dbhelper.AlterStatusDetails(utils.UnAuthorized, utils.Warned, statusDetails.AuthenticationTimes, userCredentials.ID)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "AlterNoOfTime: unable to change no of login time.")
			return
		}
		return

	case statusDetails.Type == utils.UnAuthorized && statusDetails.AuthenticationTimes > 0:

		utils.RespondJSON(w, http.StatusBadRequest, utils.ResponseMsg{
			Msg: "unauthorized email.",
		})
		DBErr := dbhelper.AlterStatusDetails(utils.Blocked, utils.Blocklisted, statusDetails.AuthenticationTimes, userCredentials.ID)
		if DBErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, DBErr, "AlterUserStatus: unable to change user status.")
			return
		}
		return

	case statusDetails.Type == utils.Blocked:

		utils.RespondJSON(w, http.StatusBadRequest, utils.ResponseMsg{
			Msg: "Blocked email.",
		})
		return
	}

	hours := 24

	expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)

	claims := &models.Claims{
		ID: userCredentials.ID,
		StandardClaims: jwt.StandardClaims{

			ExpiresAt: expiresAt.Unix(),
			// Issuer:    userCredentials.Role,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "LoginUser: cannot create tokenString.")
		return
	}

	err = dbhelper.CreateSession(claims)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "LoginUser: cannot create session.")
		return
	}

	userOutboundData := make(map[string]interface{})

	userOutboundData["token"] = tokenString

	err = utils.EncodeJSONBody(w, userOutboundData)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "LoginUser: not able to login.")
		return
	}
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	user, err := dbhelper.GetUserDetails(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "cannot get user details.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func GetUserDetails(w http.ResponseWriter, r *http.Request) {
	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "cannot get user id.")
		return
	}

	user, err := dbhelper.GetUserDetails(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "cannot get user details.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func GetDashboard(w http.ResponseWriter, r *http.Request) {
	dashBoardFilter := r.URL.Query().Get("dashBoardFilter")
	if dashBoardFilter == "" {
		dashBoardFilter = "total"
	}

	quantities, err := dbhelper.GetAssetQuantities(dashBoardFilter)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get asset quantities.")
		return
	}
	assetQuantities, err := dbhelper.GetTotalAssetQuantities(quantities)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to get total asset quantities.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, assetQuantities)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	body := models.RegisterUser{}
	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body.")
		return
	}

	validationErr := validate.Struct(body)
	if validationErr != nil {
		utils.RespondError(w, http.StatusBadRequest, validationErr, "validation error")
		return
	}

	hashedPassword, hashErr := utils.HashPassword(body.Password)
	if hashErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, hashErr, "failed to secure password.")
		return
	}

	userID, userErr := utils.UserContext(r)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "cannot get user id.")
		return
	}

	updateErr := dbhelper.UpdateUser(body, hashedPassword, userID)
	if updateErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, updateErr, "failed to update user details.")
		return
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "User details updated.",
	})
}

func UpdateAccessedBy(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	userType := r.URL.Query().Get("userType")

	err := dbhelper.UpdateAccessedBy(userID, userType)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "UpdateAccessedBy: cannot update accessedBy")
	}

	utils.RespondJSON(w, http.StatusOK, utils.ResponseMsg{
		Msg: "Updated accessed by.",
	})
}
