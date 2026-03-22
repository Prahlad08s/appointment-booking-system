package router

import (
	"appointment-booking-system/business"
	"appointment-booking-system/commons/constants"
	"appointment-booking-system/handlers"
	"appointment-booking-system/middleware"
	"appointment-booking-system/repositories"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, logger *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestLogger(logger))

	// Repositories
	coachRepo := repositories.NewCoachRepository(db)
	availRepo := repositories.NewCoachAvailabilityRepository(db)
	slotsRepo := repositories.NewSlotsRepository(db)
	bookingRepo := repositories.NewBookingRepository(db)

	// Services
	coachService := business.NewCoachService(coachRepo, logger)
	availService := business.NewCoachAvailabilityService(availRepo, logger)
	slotsService := business.NewGetAvailableSlotsService(availRepo, slotsRepo, logger)
	bookingService := business.NewCreateBookingService(bookingRepo, availRepo, logger)
	userBookingsService := business.NewGetUserBookingsService(bookingRepo, logger)
	cancelService := business.NewCancelBookingService(bookingRepo, logger)

	// Handlers
	coachHandler := handlers.NewCoachHandler(coachService)
	availHandler := handlers.NewCoachAvailabilityHandler(availService)
	slotsHandler := handlers.NewGetAvailableSlotsHandler(slotsService)
	bookingHandler := handlers.NewCreateBookingHandler(bookingService)
	userBookingsHandler := handlers.NewGetUserBookingsHandler(userBookingsService)
	cancelHandler := handlers.NewCancelBookingHandler(cancelService)

	// Routes
	r.GET(constants.HealthRoute, handlers.HealthCheck())

	r.POST(constants.CoachesRoute, coachHandler.CreateCoach())
	r.GET(constants.CoachesRoute, coachHandler.GetCoaches())

	r.POST(constants.CoachAvailabilityRoute, availHandler.SetAvailability())
	r.GET(constants.CoachAvailabilityRoute, availHandler.GetAvailability())

	r.GET(constants.UserSlotsRoute, slotsHandler.GetSlots())

	r.POST(constants.UserBookingsRoute, bookingHandler.CreateBooking())
	r.GET(constants.UserBookingsRoute, userBookingsHandler.GetBookings())
	r.DELETE(constants.CancelBookingRoute, cancelHandler.CancelBooking())

	return r
}
