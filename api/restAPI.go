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
	baseFolder = os.Getenv("BASE_INPROGRESS_FOLDER")
	ensureBaseFolderExists(baseFolder)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// GETs
	e.GET("/inprogress", inprogress)
	e.GET("/videos", videos)

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

func inprogress(c echo.Context) error {
	var inprogress []models.InProgress
	db.Find(&inprogress)

	c.Logger().Debug("Retrieved in progress")

	return c.JSON(http.StatusOK, inprogress)
}

func videos(c echo.Context) error {
	var video []models.Video
	db.Find(&video)

	c.Logger().Debug("Retrieved videos")

	return c.JSON(http.StatusOK, inprogress)
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
		c.Logger().Infof("%s: %s", errMsg, err)
		return c.String(http.StatusInternalServerError, errMsg)
	}

	var currentTracker models.InProgress
	if id == 0 {
		currentTracker, err = createNewTrackerFolder(currentTracker)
		if err != nil {
			errMsg := "Failed creating folder for progress"
			c.Logger().Infof("%s: %s", errMsg, err)
			return c.String(http.StatusInternalServerError, errMsg)
		}
	} else {
		test := db.Where(models.InProgress{Id: id}).First(&currentTracker)
		if test.Error != nil {
			errMsg := "Failed fetching progress tracker"
			c.Logger().Infof("%s: %s", errMsg, test.Error)
			return c.String(http.StatusInternalServerError, errMsg)
		}
	}

	if currentTracker.Status == "Complete" {
		currentTracker, err = createNewTrackerFolder(models.InProgress{})
		if err != nil {
			errMsg := "Failed creating folder for progress"
			c.Logger().Infof("%s: %s", errMsg, err)
			return c.String(http.StatusInternalServerError, errMsg)
		}
	}

	currentTracker.Count += 1

	newFileName := path.Join(currentTracker.Folder, fmt.Sprintf("%05d.jpg", currentTracker.Count))

	err = os.WriteFile(newFileName, image, 0644)
	if err != nil {
		errMsg := "Failed writing image"
		c.Logger().Infof("%s: %s", errMsg, err)
		return c.String(http.StatusInternalServerError, errMsg)
	}

	db.Save(currentTracker)

	c.Logger().Debugf("Added image to %d", currentTracker.Id)

	return c.String(http.StatusOK, fmt.Sprintf("%d", currentTracker.Id))
}

func createNewTrackerFolder(currentTracker models.InProgress) (models.InProgress, error) {
	currentTracker.Date = time.Now()
	newFileName := path.Join(baseFolder, uuid.New().String())
	err := os.Mkdir(newFileName, 0755)
	if err != nil {
		return currentTracker, err
	}
	currentTracker.Folder = newFileName
	db.Create(&currentTracker)

	return currentTracker, nil
}
