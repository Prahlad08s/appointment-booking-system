package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"appointment-booking-system/models"
	"appointment-booking-system/repositories"
	"appointment-booking-system/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testRouter *gin.Engine
	testLogger *zap.Logger
)

func setupTestDB() *gorm.DB {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=appointment_booking_test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	db.Exec("DROP TABLE IF EXISTS bookings")
	db.Exec("DROP TABLE IF EXISTS coach_availabilities")
	db.Exec("DROP TABLE IF EXISTS coaches")

	db.AutoMigrate(&models.Coach{}, &models.CoachAvailability{}, &models.Booking{})

	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_booking 
		ON bookings(coach_id, start_time) WHERE status = 'booked'`)

	return db
}

func setupTestRouter() *gin.Engine {
	testLogger, _ = zap.NewDevelopment()
	testDB = setupTestDB()
	repositories.DB = testDB
	testRouter = router.SetupRouter(testDB, testLogger)
	return testRouter
}

func cleanupDB() {
	testDB.Exec("DELETE FROM bookings")
	testDB.Exec("DELETE FROM coach_availabilities")
	testDB.Exec("DELETE FROM coaches")
}

func seedCoach(name, email, timezone string) models.Coach {
	coach := models.Coach{Name: name, Email: email, Timezone: timezone}
	testDB.Create(&coach)
	return coach
}

func seedAvailability(coachID uint, dayOfWeek int, startTime, endTime string) models.CoachAvailability {
	avail := models.CoachAvailability{
		CoachID:   coachID,
		DayOfWeek: dayOfWeek,
		StartTime: startTime,
		EndTime:   endTime,
	}
	testDB.Create(&avail)
	return avail
}

func performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func parseResponse(w *httptest.ResponseRecorder) models.APIResponse {
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}
