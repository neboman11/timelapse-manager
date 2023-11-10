package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/neboman11/timelapse-manager/models"
	"gorm.io/gorm"
)

var db *gorm.DB
var baseFolder string

// Starts listening for requests on the given port
func HandleRequests(port int, database *gorm.DB) {
	db = database
	baseFolder = os.Getenv("BASE_FOLDER")
	ensureBaseFolderExists(baseFolder)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// GETs
	e.GET("/timelapses", timelapses)
	e.GET("/video", video)

	// POSTs
	e.POST("/inprogress/add", add_inprogress)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func ensureBaseFolderExists(baseInProgressFolder string) {
	if _, err := os.Stat(baseInProgressFolder); err != nil {
		os.Mkdir(baseInProgressFolder, 0755)
	}
}

// Routes

// GETs

func timelapses(c echo.Context) error {
	var inprogress []models.Timelapse
	db.Find(&inprogress)

	c.Logger().Debug("Retrieved in progress")

	return c.JSON(http.StatusOK, inprogress)
}

func video(c echo.Context) error {
	idstr := c.QueryParam("id")
	if len(idstr) < 1 {
		errMsg := "Param 'id' is missing"
		c.Logger().Info(errMsg)
		return c.String(http.StatusBadRequest, errMsg)
	}
	id, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		errMsg := "Failed to parse id"
		c.Logger().Infof("%s: %s", errMsg, err)
		return c.String(http.StatusBadRequest, errMsg)
	}

	var timelapse models.Timelapse
	result := db.Where(models.Timelapse{Id: id}).First(&timelapse)
	if result.Error != nil {
		errMsg := fmt.Sprintf("Video with id %d does not exist", id)
		c.Logger().Infof("%s: %s", errMsg, result.Error)
		return c.String(http.StatusBadRequest, errMsg)
	}

	c.Logger().Debug("Retrieved video")

	return c.File(path.Join(timelapse.Folder, timelapse.VideoFile))
}

// POSTs

func add_inprogress(c echo.Context) error {
	idstr := c.QueryParam("id")
	if len(idstr) < 1 {
		errMsg := "Param 'id' is missing"
		c.Logger().Info(errMsg)
		return c.String(http.StatusBadRequest, errMsg)
	}
	id, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		errMsg := "Failed to parse id"
		c.Logger().Infof("%s: %s", errMsg, err)
		return c.String(http.StatusBadRequest, errMsg)
	}

	image, err := io.ReadAll(c.Request().Body)
	if err != nil {
		errMsg := "Failed to read request body"
		c.Logger().Errorf("%s: %s", errMsg, err)
		return c.String(http.StatusInternalServerError, errMsg)
	}

	var currentTracker models.Timelapse
	if id == 0 {
		currentTracker, err = createNewTrackerFolder(currentTracker)
		if err != nil {
			errMsg := "Failed creating folder for progress"
			c.Logger().Errorf("%s: %s", errMsg, err)
			return c.String(http.StatusInternalServerError, errMsg)
		}
	} else {
		result := db.Where(models.Timelapse{Id: id}).First(&currentTracker)
		if result.Error != nil {
			errMsg := fmt.Sprintf("Progress tracker with id %d does not exist", id)
			c.Logger().Infof("%s: %s", errMsg, result.Error)
			return c.String(http.StatusBadRequest, errMsg)
		}
	}

	if currentTracker.Status == "Complete" {
		currentTracker, err = createNewTrackerFolder(models.Timelapse{})
		if err != nil {
			errMsg := "Failed creating folder for progress"
			c.Logger().Errorf("%s: %s", errMsg, err)
			return c.String(http.StatusInternalServerError, errMsg)
		}
	}

	currentTracker.ImageCount += 1

	newFileName := path.Join(currentTracker.Folder, fmt.Sprintf("%05d.jpg", currentTracker.ImageCount))

	err = os.WriteFile(newFileName, image, 0644)
	if err != nil {
		errMsg := "Failed writing image"
		c.Logger().Errorf("%s: %s", errMsg, err)
		return c.String(http.StatusInternalServerError, errMsg)
	}

	db.Save(currentTracker)

	c.Logger().Debugf("Added image to %d", currentTracker.Id)

	return c.String(http.StatusOK, fmt.Sprintf("%d", currentTracker.Id))
}

func createNewTrackerFolder(currentTracker models.Timelapse) (models.Timelapse, error) {
	currentTracker.StartDate = time.Now()
	newFileName := path.Join(baseFolder, uuid.New().String())
	err := os.Mkdir(newFileName, 0755)
	if err != nil {
		return currentTracker, err
	}
	currentTracker.Folder = newFileName
	db.Create(&currentTracker)

	return currentTracker, nil
}
