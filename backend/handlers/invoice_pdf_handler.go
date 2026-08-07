package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"am-keramika-backend/config"
	"am-keramika-backend/internal/invoicepdf"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetInvoicePDF(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nije validan"})
		return
	}

	invoice, err := repositories.GetInvoiceByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Račun nije pronađen"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Račun nije pronađen"})
		return
	}

	mapped := mapInvoiceResponse(*invoice)
	company := config.LoadCompanyConfig()

	doc := invoicepdf.Document{
		ID:              mapped.ID,
		CreatedAt:       mapped.CreatedAt,
		Status:          mapped.Status,
		TotalAmount:     mapped.TotalAmount,
		PaidAmount:      mapped.PaidAmount,
		RemainingAmount: mapped.RemainingAmount,
		IsCashSale:      mapped.CustomerID == nil && mapped.Customer == nil,
		Company: invoicepdf.Company{
			Name:               company.Name,
			Address:            company.Address,
			City:               company.City,
			Phone:              company.Phone,
			Email:              company.Email,
			TaxID:              company.TaxID,
			RegistrationNumber: company.RegistrationNumber,
			BankAccount:        company.BankAccount,
			Website:            company.Website,
		},
	}

	if mapped.Customer != nil {
		doc.CustomerName = mapped.Customer.Name
		doc.CustomerPhone = mapped.Customer.Phone
	}
	if mapped.CreatedByUser != nil {
		doc.CreatedBy = userDisplayName(
			mapped.CreatedByUser.FullName,
			mapped.CreatedByUser.Username,
		)
	}

	doc.Items = make([]invoicepdf.Item, 0, len(mapped.Items))
	for _, item := range mapped.Items {
		doc.Items = append(doc.Items, invoicepdf.Item{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		})
	}

	pdfBytes, err := invoicepdf.Generate(doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Generisanje PDF-a nije uspjelo"})
		return
	}

	filename := invoicepdf.Filename(mapped.ID)
	dispositionType := "inline"
	if strings.EqualFold(c.Query("download"), "true") || c.Query("download") == "1" {
		dispositionType = "attachment"
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`%s; filename="%s"`, dispositionType, filename))
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
