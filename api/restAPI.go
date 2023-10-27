package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/neboman11/timelapse-manager/models"
	"gorm.io/gorm"
)

var db *gorm.DB
var baseInProgressFolder string

// Starts listening for requests on the given port
func HandleRequests(port int, database *gorm.DB) {
	db = database
	baseInProgressFolder = os.Getenv("BASE_INPROGRESS_FOLDER")
	ensureInProgressFolderExists(baseInProgressFolder)

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

func ensureInProgressFolderExists(baseInProgressFolder string) {
	if _, err := os.Stat(baseInProgressFolder); err != nil {
		os.Mkdir(baseInProgressFolder, 0755)
	}
}

// Routes

// GETs

func inprogress(c echo.Context) error {
	var inprogress []models.InProgress
	db.Find(&inprogress)

	log.Trace("Retrieved in progress")

	return c.JSON(http.StatusOK, inprogress)
}

// POSTs

func add_inprogress(c echo.Context) error {
	idstr := c.QueryParam("id")
	if len(idstr) < 1 {
		errMsg := "Param 'id' is missing"
		log.Info(errMsg)
		return c.String(http.StatusBadRequest, errMsg)
	}
	id, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		errMsg := "Failed to parse id"
		log.Infof("%s: %s", errMsg, err)
		return c.String(http.StatusBadRequest, errMsg)
	}

	image, err := io.ReadAll(c.Request().Body)
	if err != nil {
		errMsg := "Failed to read request body"
		log.Infof("%s: %s", errMsg, err)
		return c.String(http.StatusInternalServerError, errMsg)
	}

	var currentTracker models.InProgress
	if id == 0 {
		currentTracker.Date = time.Now()
		newFileName := path.Join(baseInProgressFolder, uuid.New().String())
		err = os.Mkdir(newFileName, 0755)
		if err != nil {
			errMsg := "Failed creating folder for progress"
			log.Infof("%s: %s", errMsg, err)
			return c.String(http.StatusInternalServerError, errMsg)
		}
		currentTracker.Folder = newFileName
		db.Create(&currentTracker)
	} else {
		test := db.Where(models.InProgress{Id: id}).First(&currentTracker)
		if test.Error != nil {
			errMsg := "Failed fetching progress tracker"
			log.Infof("%s: %s", errMsg, test.Error)
			return c.String(http.StatusInternalServerError, errMsg)
		}
	}

	newFileName := path.Join(currentTracker.Folder, fmt.Sprintf("%s.jpg", time.Now().Format("2006-01-02-03-04-05.999")))

	err = os.WriteFile(newFileName, image, 0644)
	if err != nil {
		errMsg := "Failed writing image"
		log.Infof("%s: %s", errMsg, err)
		return c.String(http.StatusInternalServerError, errMsg)
	}

	log.Tracef("Added image to %d", currentTracker.Id)

	return c.String(http.StatusOK, fmt.Sprintf("%d", currentTracker.Id))
}
