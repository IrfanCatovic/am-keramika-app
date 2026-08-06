package handlers

import (
	"am-keramika-backend/dto"
	"am-keramika-backend/models"
)

func mapInvoiceResponse(invoice models.Invoice) dto.InvoiceResponse {
	response := dto.InvoiceResponse{
		ID:              invoice.ID,
		CustomerID:      invoice.CustomerID,
		TotalAmount:     invoice.TotalAmount,
		PaidAmount:      invoice.PaidAmount,
		RemainingAmount: invoice.TotalAmount - invoice.PaidAmount,
		Status:          string(invoice.Status),
		CreatedAt:       invoice.CreatedAt.Format("2006-01-02 15:04"),
		Items:           make([]dto.InvoiceItemResponse, 0, len(invoice.Items)),
	}

	if invoice.Customer != nil {
		mapped := mapCustomerResponse(*invoice.Customer)
		response.Customer = &mapped
	}

	if invoice.CreatedByUser.ID != 0 {
		response.CreatedByUser = &dto.UserSummaryResponse{
			ID:       invoice.CreatedByUser.ID,
			Username: invoice.CreatedByUser.Username,
		}
	}

	for _, item := range invoice.Items {
		productName := ""
		if item.Product.ID != 0 {
			productName = item.Product.Name
		}
		response.Items = append(response.Items, dto.InvoiceItemResponse{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		})
	}

	return response
}

func mapInvoiceListResponse(invoice models.Invoice) dto.InvoiceListResponse {
	response := dto.InvoiceListResponse{
		ID:              invoice.ID,
		CustomerID:      invoice.CustomerID,
		TotalAmount:     invoice.TotalAmount,
		PaidAmount:      invoice.PaidAmount,
		RemainingAmount: invoice.TotalAmount - invoice.PaidAmount,
		Status:          string(invoice.Status),
		CreatedAt:       invoice.CreatedAt.Format("2006-01-02 15:04"),
	}

	if invoice.Customer != nil {
		mapped := mapCustomerResponse(*invoice.Customer)
		response.Customer = &mapped
		response.CustomerName = invoice.Customer.Name
	}

	if invoice.CreatedByUser.ID != 0 {
		response.CreatedByUser = &dto.UserSummaryResponse{
			ID:       invoice.CreatedByUser.ID,
			Username: invoice.CreatedByUser.Username,
		}
	}

	return response
}
