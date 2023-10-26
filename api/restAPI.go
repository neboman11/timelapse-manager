package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/neboman11/timelapse-manager/models"
	"gorm.io/gorm"
)

var db *gorm.DB

// Starts listening for requests on the given port
func HandleRequests(port int, database *gorm.DB) {
	db = database
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// GETs
	e.GET("/inprogress", inprogress)

	// POSTs
	e.POST("/inprogress/add", add_inprogress)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

// Routes

// GETs

func inprogress(c echo.Context) error {
	var inprogress []models.InProgress
	db.Find(&inprogress)
	return c.JSON(http.StatusOK, inprogress)
}

// POSTs

func add_inprogress(c echo.Context) error {
	idstr := c.QueryParam("id")
	if len(idstr) < 1 {
		return c.String(http.StatusBadRequest, "Param 'id' is missing")
	}
	id, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse id")
	}

	image, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to read request body")
	}

	var currentTracker models.InProgress
	db.First(currentTracker, &models.InProgress{Id: id})

	newFileName := path.Join(currentTracker.Folder, fmt.Sprintf("%s.jpg", time.Now().Format("2006-01-02-03-04-05.678")))

	os.WriteFile(newFileName, image, 0644)

	return c.String(http.StatusOK, "Image added")
}
