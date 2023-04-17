package server

import (
	"InternalAssetManagement/handler"
	"InternalAssetManagement/middlewares"
	"InternalAssetManagement/utils"
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	chi.Router
	server *http.Server
}

const (
	readTimeout       = 5 * time.Minute
	readHeaderTimeout = 30 * time.Second
	writeTimeout      = 5 * time.Minute
)

func SetupRoutes() *Server {
	router := chi.NewRouter()
	// router.Use(middlewares.CommonMiddlewares()...)

	router.Route("/asset-management", func(v1 chi.Router) {
		v1.Use(middlewares.CommonMiddlewares()...)
		v1.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.RespondJSON(w, http.StatusOK, struct {
				Status string `json:"status"`
			}{Status: "server is running!"})
		})
		v1.Route("/", func(public chi.Router) {
			public.Post("/login", handler.LoginUser)
		})
		v1.Route("/user", func(user chi.Router) {
			user.Use(middlewares.AuthMiddleware)
			user.Post("/register", handler.RegisterUser)
			user.Get("/info", handler.GetUserDetails)
			user.Get("/{userID}", handler.GetUserInfo)
			user.Put("/info", handler.UpdateUser)
			user.Get("/accessed-by", handler.AccessedByDetails)
			user.Put("/accessed-by", handler.UpdateAccessedBy)
			user.Get("/dashboard", handler.GetDashboard)
			user.Put("/image", handler.AddProfileImage)
			user.Route("/employee", func(employee chi.Router) {
				employee.Group(employeeRoutes)
			})
			user.Route("/asset", func(asset chi.Router) {
				asset.Group(assetRoutes)
			})
			user.Put("/log-out", handler.Logout)
		})
	})
	return &Server{
		Router: router,
	}
}

func (svc *Server) Run(port string) error {
	svc.server = &http.Server{
		Addr:              port,
		Handler:           svc.Router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}
	return svc.server.ListenAndServe()
}

func (svc *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return svc.server.Shutdown(ctx)
}
