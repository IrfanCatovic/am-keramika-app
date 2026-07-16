package handlers

import (
	"github.com/gin-gonic/gin"
	"am-keramika-backend/repositories"
	"net/http"
	"time"
)

func GetDailyReport(c *gin.Context){

	dateParam := c.Query("date")
	if dateParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datum je obavezan"})
		return
	}

	location, err := time.LoadLocation("Europe/Belgrade")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greška pri učitavanju vremenske zone"})
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02", dateParam, location)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format datuma mora biti u formatu YYYY-MM-DD"})
		return
	}
	endDate := startDate.AddDate(0, 0, 1)	
	report, err := repositories.GetDailyReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greška pri dobavljanju dnevnog izvestaja"})
		return
	}

	c.JSON(http.StatusOK, report)
}