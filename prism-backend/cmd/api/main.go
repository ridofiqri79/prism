package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ridofiqri79/prism-backend/internal/config"
	"github.com/ridofiqri79/prism-backend/internal/database"
	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/handler"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/service"
	"github.com/ridofiqri79/prism-backend/internal/sse"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database pool")
	}
	defer pool.Close()

	q := queries.New(pool)
	middleware.SetPermissionChecker(middleware.NewDatabasePermissionChecker(q))

	broker := sse.NewBroker()
	go broker.Run()

	authService := service.NewAuthService(q, cfg.JWTSecret, cfg.JWTExpiresIn)
	userService := service.NewUserService(pool, q)
	masterService := service.NewMasterService(pool, q)
	blueBookService := service.NewBlueBookService(pool, q, broker)
	greenBookService := service.NewGreenBookService(pool, q, broker)
	dkService := service.NewDKService(pool, q, broker)
	laService := service.NewLAService(pool, q, broker)
	monitoringService := service.NewMonitoringService(pool, q, broker)
	dashboardService := service.NewDashboardService(q)
	journeyService := service.NewJourneyService(q)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	masterHandler := handler.NewMasterHandler(masterService)
	blueBookHandler := handler.NewBlueBookHandler(blueBookService)
	greenBookHandler := handler.NewGreenBookHandler(greenBookService)
	dkHandler := handler.NewDKHandler(dkService)
	laHandler := handler.NewLAHandler(laService)
	monitoringHandler := handler.NewMonitoringHandler(monitoringService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	journeyHandler := handler.NewJourneyHandler(journeyService)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = middleware.ErrorHandler

	e.Use(echomiddleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	e.POST("/api/v1/auth/login", authHandler.Login)

	api := e.Group("/api/v1")
	api.Use(middleware.Auth(cfg.JWTSecret))
	api.Use(middleware.SetAuditUser(pool))

	authGroup := api.Group("/auth")
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.GET("/me", authHandler.Me)

	userGroup := api.Group("/users", middleware.RequireAdmin())
	userGroup.GET("", userHandler.List)
	userGroup.POST("", userHandler.Create)
	userGroup.GET("/:id", userHandler.Get)
	userGroup.PUT("/:id", userHandler.Update)
	userGroup.DELETE("/:id", userHandler.Delete)
	userGroup.GET("/:id/permissions", userHandler.GetPermissions)
	userGroup.PUT("/:id/permissions", userHandler.UpdatePermissions)

	master := api.Group("/master")
	master.GET("/countries", masterHandler.ListCountries, middleware.Require("country", "read"))
	master.GET("/countries/:id", masterHandler.GetCountry, middleware.Require("country", "read"))
	master.POST("/countries", masterHandler.CreateCountry, middleware.Require("country", "create"))
	master.PUT("/countries/:id", masterHandler.UpdateCountry, middleware.Require("country", "update"))
	master.DELETE("/countries/:id", masterHandler.DeleteCountry, middleware.Require("country", "delete"))

	master.GET("/lenders", masterHandler.ListLenders, middleware.Require("lender", "read"))
	master.GET("/lenders/:id", masterHandler.GetLender, middleware.Require("lender", "read"))
	master.POST("/lenders", masterHandler.CreateLender, middleware.Require("lender", "create"))
	master.PUT("/lenders/:id", masterHandler.UpdateLender, middleware.Require("lender", "update"))
	master.DELETE("/lenders/:id", masterHandler.DeleteLender, middleware.Require("lender", "delete"))

	master.GET("/institutions", masterHandler.ListInstitutions, middleware.Require("institution", "read"))
	master.GET("/institutions/:id", masterHandler.GetInstitution, middleware.Require("institution", "read"))
	master.POST("/institutions", masterHandler.CreateInstitution, middleware.Require("institution", "create"))
	master.PUT("/institutions/:id", masterHandler.UpdateInstitution, middleware.Require("institution", "update"))
	master.DELETE("/institutions/:id", masterHandler.DeleteInstitution, middleware.Require("institution", "delete"))

	master.GET("/regions", masterHandler.ListRegions, middleware.Require("region", "read"))
	master.GET("/regions/:id", masterHandler.GetRegion, middleware.Require("region", "read"))
	master.POST("/regions", masterHandler.CreateRegion, middleware.Require("region", "create"))
	master.PUT("/regions/:id", masterHandler.UpdateRegion, middleware.Require("region", "update"))
	master.DELETE("/regions/:id", masterHandler.DeleteRegion, middleware.Require("region", "delete"))

	master.GET("/program-titles", masterHandler.ListProgramTitles, middleware.Require("program_title", "read"))
	master.GET("/program-titles/:id", masterHandler.GetProgramTitle, middleware.Require("program_title", "read"))
	master.POST("/program-titles", masterHandler.CreateProgramTitle, middleware.Require("program_title", "create"))
	master.PUT("/program-titles/:id", masterHandler.UpdateProgramTitle, middleware.Require("program_title", "update"))
	master.DELETE("/program-titles/:id", masterHandler.DeleteProgramTitle, middleware.Require("program_title", "delete"))

	master.GET("/bappenas-partners", masterHandler.ListBappenasPartners, middleware.Require("bappenas_partner", "read"))
	master.GET("/bappenas-partners/:id", masterHandler.GetBappenasPartner, middleware.Require("bappenas_partner", "read"))
	master.POST("/bappenas-partners", masterHandler.CreateBappenasPartner, middleware.Require("bappenas_partner", "create"))
	master.PUT("/bappenas-partners/:id", masterHandler.UpdateBappenasPartner, middleware.Require("bappenas_partner", "update"))
	master.DELETE("/bappenas-partners/:id", masterHandler.DeleteBappenasPartner, middleware.Require("bappenas_partner", "delete"))

	master.GET("/periods", masterHandler.ListPeriods, middleware.Require("period", "read"))
	master.GET("/periods/:id", masterHandler.GetPeriod, middleware.Require("period", "read"))
	master.POST("/periods", masterHandler.CreatePeriod, middleware.Require("period", "create"))
	master.PUT("/periods/:id", masterHandler.UpdatePeriod, middleware.Require("period", "update"))
	master.DELETE("/periods/:id", masterHandler.DeletePeriod, middleware.Require("period", "delete"))

	master.GET("/national-priorities", masterHandler.ListNationalPriorities, middleware.Require("national_priority", "read"))
	master.GET("/national-priorities/:id", masterHandler.GetNationalPriority, middleware.Require("national_priority", "read"))
	master.POST("/national-priorities", masterHandler.CreateNationalPriority, middleware.Require("national_priority", "create"))
	master.PUT("/national-priorities/:id", masterHandler.UpdateNationalPriority, middleware.Require("national_priority", "update"))
	master.DELETE("/national-priorities/:id", masterHandler.DeleteNationalPriority, middleware.Require("national_priority", "delete"))
	master.GET("/import-data/template", masterHandler.DownloadImportTemplate, middleware.RequireAdmin())
	master.POST("/import-data/preview", masterHandler.PreviewImportData, middleware.RequireAdmin())
	master.POST("/import-data/execute", masterHandler.ImportData, middleware.RequireAdmin())
	master.POST("/import-data", masterHandler.ImportData, middleware.RequireAdmin())

	blueBooks := api.Group("/blue-books")
	blueBooks.GET("", blueBookHandler.ListBlueBooks, middleware.Require("blue_book", "read"))
	blueBooks.POST("", blueBookHandler.CreateBlueBook, middleware.Require("blue_book", "create"))
	blueBooks.GET("/:id", blueBookHandler.GetBlueBook, middleware.Require("blue_book", "read"))
	blueBooks.PUT("/:id", blueBookHandler.UpdateBlueBook, middleware.Require("blue_book", "update"))
	blueBooks.DELETE("/:id", blueBookHandler.DeleteBlueBook, middleware.Require("blue_book", "delete"))

	blueBooks.GET("/:bbId/projects", blueBookHandler.ListBBProjects, middleware.Require("bb_project", "read"))
	blueBooks.POST("/:bbId/projects", blueBookHandler.CreateBBProject, middleware.Require("bb_project", "create"))
	blueBooks.GET("/:bbId/projects/:id", blueBookHandler.GetBBProject, middleware.Require("bb_project", "read"))
	blueBooks.PUT("/:bbId/projects/:id", blueBookHandler.UpdateBBProject, middleware.Require("bb_project", "update"))
	blueBooks.DELETE("/:bbId/projects/:id", blueBookHandler.DeleteBBProject, middleware.Require("bb_project", "delete"))
	blueBooks.POST("/:bbId/import-projects/preview", blueBookHandler.PreviewImportBBProjects, middleware.RequireAdmin())
	blueBooks.POST("/:bbId/import-projects/execute", blueBookHandler.ImportBBProjects, middleware.RequireAdmin())
	blueBooks.GET("/:bbId/import-projects/template", blueBookHandler.DownloadBBProjectImportTemplate, middleware.RequireAdmin())

	loi := api.Group("/bb-projects/:bbProjectId/loi")
	loi.GET("", blueBookHandler.ListLoI, middleware.Require("bb_project", "read"))
	loi.POST("", blueBookHandler.CreateLoI, middleware.Require("bb_project", "update"))
	loi.DELETE("/:id", blueBookHandler.DeleteLoI, middleware.Require("bb_project", "update"))

	greenBooks := api.Group("/green-books")
	greenBooks.GET("", greenBookHandler.ListGreenBooks, middleware.Require("green_book", "read"))
	greenBooks.POST("", greenBookHandler.CreateGreenBook, middleware.Require("green_book", "create"))
	greenBooks.GET("/:id", greenBookHandler.GetGreenBook, middleware.Require("green_book", "read"))
	greenBooks.PUT("/:id", greenBookHandler.UpdateGreenBook, middleware.Require("green_book", "update"))
	greenBooks.DELETE("/:id", greenBookHandler.DeleteGreenBook, middleware.Require("green_book", "delete"))

	greenBooks.GET("/:gbId/projects", greenBookHandler.ListGBProjects, middleware.Require("gb_project", "read"))
	greenBooks.POST("/:gbId/projects", greenBookHandler.CreateGBProject, middleware.Require("gb_project", "create"))
	greenBooks.GET("/:gbId/projects/:id", greenBookHandler.GetGBProject, middleware.Require("gb_project", "read"))
	greenBooks.PUT("/:gbId/projects/:id", greenBookHandler.UpdateGBProject, middleware.Require("gb_project", "update"))
	greenBooks.DELETE("/:gbId/projects/:id", greenBookHandler.DeleteGBProject, middleware.Require("gb_project", "delete"))

	dk := api.Group("/daftar-kegiatan")
	dk.GET("", dkHandler.ListDK, middleware.Require("daftar_kegiatan", "read"))
	dk.POST("", dkHandler.CreateDK, middleware.Require("daftar_kegiatan", "create"))
	dk.GET("/:id", dkHandler.GetDK, middleware.Require("daftar_kegiatan", "read"))
	dk.PUT("/:id", dkHandler.UpdateDK, middleware.Require("daftar_kegiatan", "update"))
	dk.DELETE("/:id", dkHandler.DeleteDK, middleware.Require("daftar_kegiatan", "delete"))

	dk.GET("/:dkId/projects", dkHandler.ListDKProjects, middleware.Require("daftar_kegiatan", "read"))
	dk.POST("/:dkId/projects", dkHandler.CreateDKProject, middleware.Require("daftar_kegiatan", "create"))
	dk.GET("/:dkId/projects/:id", dkHandler.GetDKProject, middleware.Require("daftar_kegiatan", "read"))
	dk.PUT("/:dkId/projects/:id", dkHandler.UpdateDKProject, middleware.Require("daftar_kegiatan", "update"))
	dk.DELETE("/:dkId/projects/:id", dkHandler.DeleteDKProject, middleware.Require("daftar_kegiatan", "delete"))

	loanAgreements := api.Group("/loan-agreements")
	loanAgreements.GET("", laHandler.ListLA, middleware.Require("loan_agreement", "read"))
	loanAgreements.POST("", laHandler.CreateLA, middleware.Require("loan_agreement", "create"))
	loanAgreements.GET("/:id", laHandler.GetLA, middleware.Require("loan_agreement", "read"))
	loanAgreements.PUT("/:id", laHandler.UpdateLA, middleware.Require("loan_agreement", "update"))
	loanAgreements.DELETE("/:id", laHandler.DeleteLA, middleware.Require("loan_agreement", "delete"))

	monitoring := api.Group("/loan-agreements/:laId/monitoring")
	monitoring.GET("", monitoringHandler.List, middleware.Require("monitoring_disbursement", "read"))
	monitoring.POST("", monitoringHandler.Create, middleware.Require("monitoring_disbursement", "create"))
	monitoring.GET("/:id", monitoringHandler.Get, middleware.Require("monitoring_disbursement", "read"))
	monitoring.PUT("/:id", monitoringHandler.Update, middleware.Require("monitoring_disbursement", "update"))
	monitoring.DELETE("/:id", monitoringHandler.Delete, middleware.Require("monitoring_disbursement", "delete"))

	dashboard := api.Group("/dashboard")
	dashboard.GET("/summary", dashboardHandler.Summary)
	dashboard.GET("/monitoring-summary", dashboardHandler.MonitoringSummary)

	api.GET("/projects/:bbProjectId/journey", journeyHandler.GetJourney, middleware.Require("bb_project", "read"))

	e.GET("/events", handler.SSEHandler(broker))

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	waitForShutdown(e)
}

func waitForShutdown(e *echo.Echo) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shut down server gracefully")
	}
}
