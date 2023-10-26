package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

var db *gorm.DB

var makingRequest sync.Mutex

// Starts listening for requests on the given port
func HandleRequests(port int, database *gorm.DB) {
	db = database
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// GETs
	e.GET("/wanted", wanted)
	e.GET("/cover", cover)

	// DELETEs
	e.DELETE("/delete", delete)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

// Routes

// GETs

func wanted(c echo.Context) error {
	var want []Want
	db.Find(&want)
	return c.JSON(http.StatusOK, want)
}

func cover(c echo.Context) error {
	artist := c.QueryParam("artist")
	if len(artist) < 1 {
		return c.String(http.StatusBadRequest, "Param 'artist' is missing")
	}

	album := c.QueryParam("album")
	if len(album) < 1 {
		return c.String(http.StatusBadRequest, "Param 'album' is missing")
	}

	musicbrainz_ids, err := get_musicbrainz_ids(artist, album)
	if err != nil {
		log.Errorf("Failed to get musicbrainz id: %s", err)
		return c.String(http.StatusNotFound, "Failed to get MusicBrainz ID")
	}

	album_art_link, err := get_album_art_link(musicbrainz_ids)
	if err != nil {
		log.Errorf("Failed to get album art link: %s", err)
		return c.String(http.StatusNotFound, "Failed to get album art link")
	}

	return c.JSON(http.StatusOK, CoverResponse{album_art_link})
}

// DELETEs

func delete(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to read request body")
	}

	var albums []Album
	if err := json.Unmarshal(body, &albums); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to parse request body")
	}

	db.Delete(&Want{}, albums)

	return c.String(http.StatusOK, "Albums deleted")
}

// Private Functions

func get_musicbrainz_ids(artist string, album string) ([]string, error) {
	makingRequest.Lock()

	resp, err := http.Get("https://musicbrainz.org/ws/2/release/?query=" + url.QueryEscape(fmt.Sprintf("artistname:%s AND release:%s", artist, album)) + "&fmt=json")
	time.Sleep(100 * time.Millisecond)
	makingRequest.Unlock()
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get musicbrainz id: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mbResp MusicBrainzResponse
	if err := json.Unmarshal(body, &mbResp); err != nil {
		return nil, err
	}

	if len(mbResp.Releases) < 1 {
		return nil, fmt.Errorf("no releases found for %s - %s\nbody: %s", artist, album, body)
	}

	ids := make([]string, len(mbResp.Releases))

	for i, release := range mbResp.Releases {
		ids[i] = release.Id
	}

	return ids, nil
}

func get_album_art_link(musicbrainz_ids []string) (string, error) {
	for _, id := range musicbrainz_ids {
		link, err := sub_get_album_art_link(id)
		if err != nil {
			return "", err
		}
		if len(link) > 0 {
			return link, nil
		}
	}
	return "", fmt.Errorf("no images found")
}

func sub_get_album_art_link(musicbrainz_id string) (string, error) {
	makingRequest.Lock()

	resp, err := http.Get(fmt.Sprintf("https://coverartarchive.org/release/%s?fmt=json", musicbrainz_id))
	time.Sleep(100 * time.Millisecond)
	makingRequest.Unlock()
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get album art link: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var coverArtResponse struct {
		Images []struct {
			Image string `json:"image"`
			Front bool   `json:"front"`
		} `json:"images"`
	}
	if err := json.Unmarshal(body, &coverArtResponse); err != nil {
		return "", err
	}

	if len(coverArtResponse.Images) > 0 && coverArtResponse.Images[0].Front {
		return coverArtResponse.Images[0].Image, nil
	}

	return "", nil
}
