package repositories

import (
	"testing"
	"time"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupInvoiceListTestDB(t *testing.T) {
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
		&models.InvoiceItem{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedInvoiceListUser(t *testing.T) models.User {
	t.Helper()
	user := models.User{Username: "sef", PasswordHash: "x", Role: models.RoleBoss, IsActive: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	return user
}

func seedInvoiceAt(t *testing.T, userID uint, customerID *uint, status models.InvoiceStatus, createdAt time.Time, total float64) models.Invoice {
	t.Helper()
	invoice := models.Invoice{
		CreatedByUserID: userID,
		CustomerID:      customerID,
		Status:          status,
		TotalAmount:     total,
		PaidAmount:      0,
	}
	if err := database.DB.Create(&invoice).Error; err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if err := database.DB.Model(&invoice).Update("created_at", createdAt).Error; err != nil {
		t.Fatalf("set created_at: %v", err)
	}
	invoice.CreatedAt = createdAt
	return invoice
}

func belgradeDay(t *testing.T, date string) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Belgrade")
	if err != nil {
		t.Fatalf("location: %v", err)
	}
	day, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return day
}

func TestGetAllInvoicesFilterByCustomerID(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)
	c1 := models.Customer{Name: "Dallas", Phone: "061", IsActive: true}
	c2 := models.Customer{Name: "Drugi", Phone: "062", IsActive: true}
	database.DB.Create(&c1)
	database.DB.Create(&c2)

	day := belgradeDay(t, "2026-08-10").Add(12 * time.Hour)
	seedInvoiceAt(t, user.ID, &c1.ID, models.InvoiceStatusUnpaid, day, 100)
	seedInvoiceAt(t, user.ID, &c2.ID, models.InvoiceStatusUnpaid, day, 200)

	invoices, total, err := GetAllInvoices(InvoiceListQuery{Page: 1, Limit: 20, CustomerID: "1"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(invoices) != 1 || invoices[0].CustomerID == nil || *invoices[0].CustomerID != c1.ID {
		t.Fatalf("expected customer 1 only, got total=%d %+v", total, invoices)
	}
}

func TestGetAllInvoicesFromDateOnly(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)

	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, belgradeDay(t, "2026-07-31").Add(12*time.Hour), 10)
	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, belgradeDay(t, "2026-08-01").Add(12*time.Hour), 20)

	from := belgradeDay(t, "2026-08-01")
	invoices, total, err := GetAllInvoices(InvoiceListQuery{Page: 1, Limit: 20, FromDate: &from})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(invoices) != 1 || invoices[0].TotalAmount != 20 {
		t.Fatalf("expected only Aug 1+, got total=%d %+v", total, invoices)
	}
}

func TestGetAllInvoicesToDateOnly(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)

	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, belgradeDay(t, "2026-08-31").Add(23*time.Hour), 30)
	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, belgradeDay(t, "2026-09-01").Add(time.Hour), 40)

	toExclusive := belgradeDay(t, "2026-08-31").AddDate(0, 0, 1)
	invoices, total, err := GetAllInvoices(InvoiceListQuery{Page: 1, Limit: 20, ToDate: &toExclusive})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(invoices) != 1 || invoices[0].TotalAmount != 30 {
		t.Fatalf("expected only through Aug 31, got total=%d %+v", total, invoices)
	}
}

func TestGetAllInvoicesCombinedPeriod(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)

	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, belgradeDay(t, "2026-07-31").Add(12*time.Hour), 10)
	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, belgradeDay(t, "2026-08-15").Add(12*time.Hour), 20)
	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, belgradeDay(t, "2026-09-01").Add(time.Hour), 30)

	from := belgradeDay(t, "2026-08-01")
	toExclusive := belgradeDay(t, "2026-08-31").AddDate(0, 0, 1)
	invoices, total, err := GetAllInvoices(InvoiceListQuery{Page: 1, Limit: 20, FromDate: &from, ToDate: &toExclusive})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(invoices) != 1 || invoices[0].TotalAmount != 20 {
		t.Fatalf("expected Aug period only, got total=%d %+v", total, invoices)
	}
}

func TestGetAllInvoicesCombinedStatusCustomerPeriod(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)
	customer := models.Customer{Name: "Dallas", Phone: "061", IsActive: true}
	database.DB.Create(&customer)
	other := models.Customer{Name: "Other", Phone: "062", IsActive: true}
	database.DB.Create(&other)

	day := belgradeDay(t, "2026-08-10").Add(12 * time.Hour)
	seedInvoiceAt(t, user.ID, &customer.ID, models.InvoiceStatusPartiallyPaid, day, 100)
	seedInvoiceAt(t, user.ID, &customer.ID, models.InvoiceStatusUnpaid, day, 200)
	seedInvoiceAt(t, user.ID, &other.ID, models.InvoiceStatusPartiallyPaid, day, 300)
	seedInvoiceAt(t, user.ID, &customer.ID, models.InvoiceStatusPartiallyPaid, belgradeDay(t, "2026-07-01").Add(12*time.Hour), 400)

	from := belgradeDay(t, "2026-08-01")
	toExclusive := belgradeDay(t, "2026-08-31").AddDate(0, 0, 1)
	invoices, total, err := GetAllInvoices(InvoiceListQuery{
		Page:       1,
		Limit:      20,
		Status:     string(models.InvoiceStatusPartiallyPaid),
		CustomerID: "1",
		FromDate:   &from,
		ToDate:     &toExclusive,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(invoices) != 1 || invoices[0].TotalAmount != 100 {
		t.Fatalf("expected one combined match, got total=%d %+v", total, invoices)
	}
	if string(models.InvoiceStatusPartiallyPaid) != "partially_paid" {
		t.Fatalf("expected partially_paid constant, got %q", models.InvoiceStatusPartiallyPaid)
	}
}

func TestGetAllInvoicesSearchByCustomerName(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)
	customer := models.Customer{Name: "Dallas Shop", Phone: "061", IsActive: true}
	database.DB.Create(&customer)
	day := belgradeDay(t, "2026-08-10").Add(12 * time.Hour)
	seedInvoiceAt(t, user.ID, &customer.ID, models.InvoiceStatusUnpaid, day, 100)
	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, day, 50)

	invoices, total, err := GetAllInvoices(InvoiceListQuery{Page: 1, Limit: 20, Search: "  dallas "})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(invoices) != 1 {
		t.Fatalf("expected name search match, got total=%d %+v", total, invoices)
	}
}

func TestGetAllInvoicesSearchByInvoiceID(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)
	day := belgradeDay(t, "2026-08-10").Add(12 * time.Hour)
	inv := seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, day, 50)
	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, day, 60)

	invoices, total, err := GetAllInvoices(InvoiceListQuery{Page: 1, Limit: 20, Search: "1"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(invoices) != 1 || invoices[0].ID != inv.ID {
		t.Fatalf("expected invoice id search, got total=%d %+v", total, invoices)
	}
}

func TestGetAllInvoicesKeepsCashSalesInDefaultList(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)
	customer := models.Customer{Name: "Kupac", Phone: "061", IsActive: true}
	database.DB.Create(&customer)
	day := belgradeDay(t, "2026-08-10").Add(12 * time.Hour)
	seedInvoiceAt(t, user.ID, &customer.ID, models.InvoiceStatusUnpaid, day, 100)
	seedInvoiceAt(t, user.ID, nil, models.InvoiceStatusPaid, day, 50)

	invoices, total, err := GetAllInvoices(InvoiceListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(invoices) != 2 {
		t.Fatalf("expected cash + customer invoices, got total=%d len=%d", total, len(invoices))
	}
}

func TestGetAllInvoicesPaginationTotalUsesFilters(t *testing.T) {
	setupInvoiceListTestDB(t)
	user := seedInvoiceListUser(t)
	customer := models.Customer{Name: "Dallas", Phone: "061", IsActive: true}
	database.DB.Create(&customer)
	day := belgradeDay(t, "2026-08-10").Add(12 * time.Hour)

	for i := 0; i < 5; i++ {
		seedInvoiceAt(t, user.ID, &customer.ID, models.InvoiceStatusUnpaid, day, float64(10+i))
	}
	seedInvoiceAt(t, user.ID, &customer.ID, models.InvoiceStatusPaid, day, 99)

	_, total, err := GetAllInvoices(InvoiceListQuery{
		Page:       1,
		Limit:      2,
		Status:     string(models.InvoiceStatusUnpaid),
		CustomerID: "1",
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected filtered total 5, got %d", total)
	}

	page2, total2, err := GetAllInvoices(InvoiceListQuery{
		Page:       2,
		Limit:      2,
		Status:     string(models.InvoiceStatusUnpaid),
		CustomerID: "1",
	})
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if total2 != 5 || len(page2) != 2 {
		t.Fatalf("expected page2 len 2 total 5, got total=%d len=%d", total2, len(page2))
	}
}

func TestPartiallyPaidConstantBlocksCustomerDeactivation(t *testing.T) {
	setupInvoiceListTestDB(t)
	customer := models.Customer{Name: "Kupac", Phone: "061", IsActive: true}
	database.DB.Create(&customer)

	invoice := models.Invoice{
		CustomerID:      &customer.ID,
		CreatedByUserID: 1,
		Status:          models.InvoiceStatusPartiallyPaid,
		TotalAmount:     100,
		PaidAmount:      40,
	}
	database.DB.Create(&invoice)

	if string(invoice.Status) != "partially_paid" {
		t.Fatalf("expected stored status partially_paid, got %q", invoice.Status)
	}

	err := UpdateCustomerStatus(customer.ID, false)
	if err != ErrCustomerHasOpenInvoices {
		t.Fatalf("expected open invoices error for partially_paid, got %v", err)
	}
}
