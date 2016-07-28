package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
)

type ptrailResponse struct {
	MinID  uint64        `json:"min_id,string"`
	MaxID  uint64        `json:"max_id,string"`
	Events []ptrailEvent `json:"events"`
}

type ptrailEvent struct {
	ReceivedAt time.Time `json:"received_at"`
	Message    string    `json:"message"`
	ID         uint64    `json:"id,string"`
}

func handleAppLog(c *gin.Context) {
	cfg := c.MustGet("config").(*models.Config)
	log := c.MustGet("log").(logrus.FieldLogger)

	appName := c.Param("app")

	url := fmt.Sprintf("https://papertrailapp.com/api/v1/events/search.json?q=program:%s", appName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
	}
	req.Header.Set("X-Papertrail-Token", cfg.PapertrailToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Error("Papertrail request failed")
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrAppLog))
		return
	}
	if resp.StatusCode != 200 {
		log.WithField("http_status", resp.StatusCode).Error("Papertrail request returned bad status")
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrAppLog))
		return
	}

	var presp ptrailResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&presp)
	if err != nil {
		log.WithError(err).Error("failed to decode Papertrail response")
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrAppLog))
		return
	}

	// TODO: could reduce allocations here in a variety of ways if needed
	var buf bytes.Buffer
	for _, ev := range presp.Events {
		fmt.Fprintf(&buf, "%v: %s\n", ev.ReceivedAt, ev.Message)
	}
	appLog := &models.AppLog{
		Log: buf.String(),
	}
	c.JSON(http.StatusOK, appLog)
}
