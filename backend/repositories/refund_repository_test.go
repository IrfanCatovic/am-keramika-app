package repositories_test

import (
	"strconv"
	"testing"
	"time"

	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRefundTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Customer{},
		&models.Invoice{},
		&models.Refund{},
		&models.InvoiceCancellation{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedRefundFixture(t *testing.T) (*models.User, *models.Customer, *models.Invoice, *models.Refund) {
	t.Helper()
	user := models.User{Username: "sef1", PasswordHash: "x", Role: models.RoleBoss, IsActive: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	customer := models.Customer{Name: "Mujo", Phone: "061", IsActive: true}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("customer: %v", err)
	}
	invoice := models.Invoice{
		CreatedByUserID: user.ID,
		CustomerID:      &customer.ID,
		TotalAmount:     10000,
		PaidAmount:      10000,
		Status:          models.InvoiceStatusCancelled,
	}
	if err := database.DB.Create(&invoice).Error; err != nil {
		t.Fatalf("invoice: %v", err)
	}
	refund := models.Refund{
		InvoiceID:       invoice.ID,
		CreatedByUserID: user.ID,
		Amount:          10000,
		Reason:          "Storno",
	}
	if err := database.DB.Create(&refund).Error; err != nil {
		t.Fatalf("refund: %v", err)
	}
	return &user, &customer, &invoice, &refund
}

func TestListRefundsPaginationAndFilters(t *testing.T) {
	setupRefundTestDB(t)
	_, customer, invoice, _ := seedRefundFixture(t)

	user2 := models.User{Username: "sef2", PasswordHash: "x", Role: models.RoleBoss, IsActive: true}
	database.DB.Create(&user2)
	invoice2 := models.Invoice{
		CreatedByUserID: user2.ID,
		CustomerID:      &customer.ID,
		TotalAmount:     5000,
		PaidAmount:      5000,
		Status:          models.InvoiceStatusCancelled,
	}
	database.DB.Create(&invoice2)
	database.DB.Create(&models.Refund{
		InvoiceID:       invoice2.ID,
		CreatedByUserID: user2.ID,
		Amount:          5000,
		Reason:          "Storno 2",
	})

	page1, total, err := repositories.ListRefunds(repositories.RefundListQuery{Page: 1, Limit: 1})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(page1) != 1 {
		t.Fatalf("pagination total=%d len=%d", total, len(page1))
	}

	byInvoice, totalInv, err := repositories.ListRefunds(repositories.RefundListQuery{
		Page: 1, Limit: 20, InvoiceID: strconv.FormatUint(uint64(invoice.ID), 10),
	})
	if err != nil {
		t.Fatalf("filter invoice: %v", err)
	}
	if totalInv != 1 || byInvoice[0].InvoiceID != invoice.ID {
		t.Fatalf("invoice filter failed")
	}

	byCustomer, totalCust, err := repositories.ListRefunds(repositories.RefundListQuery{
		Page: 1, Limit: 20, CustomerID: strconv.FormatUint(uint64(customer.ID), 10),
	})
	if err != nil {
		t.Fatalf("filter customer: %v", err)
	}
	if totalCust != 2 || len(byCustomer) != 2 {
		t.Fatalf("customer filter total=%d", totalCust)
	}

	from := time.Now().AddDate(0, 0, 1)
	empty, totalEmpty, err := repositories.ListRefunds(repositories.RefundListQuery{
		Page: 1, Limit: 20, FromDate: &from,
	})
	if err != nil {
		t.Fatalf("date filter: %v", err)
	}
	if totalEmpty != 0 || len(empty) != 0 {
		t.Fatalf("expected empty future filter")
	}
}

func TestListRefundsStableSort(t *testing.T) {
	setupRefundTestDB(t)
	seedRefundFixture(t)

	refunds, _, err := repositories.ListRefunds(repositories.RefundListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(refunds) < 1 {
		t.Fatal("expected refunds")
	}
	for i := 1; i < len(refunds); i++ {
		prev, curr := refunds[i-1], refunds[i]
		if prev.CreatedAt.Before(curr.CreatedAt) {
			t.Fatalf("sort not desc by created_at")
		}
		if prev.CreatedAt.Equal(curr.CreatedAt) && prev.ID < curr.ID {
			t.Fatalf("sort not desc by id on equal dates")
		}
	}
}

func TestGetRefundByInvoiceID(t *testing.T) {
	setupRefundTestDB(t)
	_, _, invoice, refund := seedRefundFixture(t)

	got, err := repositories.GetRefundByInvoiceID(invoice.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != refund.ID || got.Amount != 10000 {
		t.Fatalf("unexpected refund %+v", got)
	}
}
