package repositories

import (
	"math"
	"strconv"
	"testing"
	"time"

	"am-keramika-backend/database"
	"am-keramika-backend/dto"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupPaymentTestDB(t *testing.T) {
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
		&models.Payment{},
		&models.PaymentAllocation{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func seedPaymentUser(t *testing.T) models.User {
	t.Helper()
	user := models.User{Username: "sef_pay", PasswordHash: "x", Role: models.RoleBoss, IsActive: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	return user
}

func seedPaymentCustomer(t *testing.T, name string, debt float64) models.Customer {
	t.Helper()
	c := models.Customer{Name: name, Phone: "061", IsActive: true, TotalDebt: debt}
	if err := database.DB.Create(&c).Error; err != nil {
		t.Fatalf("customer: %v", err)
	}
	return c
}

func seedOpenInvoice(t *testing.T, userID, customerID uint, total, paid float64, status models.InvoiceStatus) models.Invoice {
	t.Helper()
	inv := models.Invoice{
		CreatedByUserID: userID,
		CustomerID:      &customerID,
		Status:          status,
		TotalAmount:     total,
		PaidAmount:      paid,
	}
	if err := database.DB.Create(&inv).Error; err != nil {
		t.Fatalf("invoice: %v", err)
	}
	return inv
}

func TestCreatePayment_FullClosesInvoice(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Dallas", 1000)
	inv := seedOpenInvoice(t, user.ID, customer.ID, 1000, 0, models.InvoiceStatusUnpaid)

	payment, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 1000,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv.ID, Amount: 1000},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if payment.TotalAmount != 1000 {
		t.Fatalf("payment total=%v", payment.TotalAmount)
	}

	var updated models.Invoice
	database.DB.First(&updated, inv.ID)
	if updated.Status != models.InvoiceStatusPaid || updated.PaidAmount != 1000 {
		t.Fatalf("invoice=%+v", updated)
	}
	var c models.Customer
	database.DB.First(&c, customer.ID)
	if c.TotalDebt != 0 {
		t.Fatalf("debt=%v", c.TotalDebt)
	}
}

func TestCreatePayment_PartialSetsPartiallyPaid(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Partial", 500)
	inv := seedOpenInvoice(t, user.ID, customer.ID, 500, 0, models.InvoiceStatusUnpaid)

	_, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 150,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv.ID, Amount: 150},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	var updated models.Invoice
	database.DB.First(&updated, inv.ID)
	if updated.Status != models.InvoiceStatusPartiallyPaid || updated.PaidAmount != 150 {
		t.Fatalf("invoice=%+v", updated)
	}
	var c models.Customer
	database.DB.First(&c, customer.ID)
	if c.TotalDebt != 350 {
		t.Fatalf("debt=%v", c.TotalDebt)
	}
}

func TestCreatePayment_MultiInvoiceAllocates(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Multi", 800)
	inv1 := seedOpenInvoice(t, user.ID, customer.ID, 300, 0, models.InvoiceStatusUnpaid)
	inv2 := seedOpenInvoice(t, user.ID, customer.ID, 500, 0, models.InvoiceStatusUnpaid)

	payment, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 450,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv1.ID, Amount: 300},
			{InvoiceID: inv2.ID, Amount: 150},
		},
	}, user.ID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(payment.Allocations) != 2 {
		t.Fatalf("allocations=%d", len(payment.Allocations))
	}
	var a, b models.Invoice
	database.DB.First(&a, inv1.ID)
	database.DB.First(&b, inv2.ID)
	if a.Status != models.InvoiceStatusPaid || a.PaidAmount != 300 {
		t.Fatalf("inv1=%+v", a)
	}
	if b.Status != models.InvoiceStatusPartiallyPaid || b.PaidAmount != 150 {
		t.Fatalf("inv2=%+v", b)
	}
}

func TestCreatePayment_SumMismatchRejected(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Mismatch", 100)
	inv := seedOpenInvoice(t, user.ID, customer.ID, 100, 0, models.InvoiceStatusUnpaid)

	_, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 100,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv.ID, Amount: 50},
		},
	}, user.ID)
	if err == nil || err.Error() != "ukupan iznos uplate se ne poklapa sa raspodelom po racunima" {
		t.Fatalf("expected sum mismatch, got %v", err)
	}
}

func TestCreatePayment_OverRemainingRejected(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Over", 100)
	inv := seedOpenInvoice(t, user.ID, customer.ID, 100, 40, models.InvoiceStatusPartiallyPaid)

	_, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 70,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv.ID, Amount: 70},
		},
	}, user.ID)
	if err == nil {
		t.Fatal("expected remaining overflow error")
	}
}

func TestCreatePayment_DuplicateInvoiceRejected(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Dup", 200)
	inv := seedOpenInvoice(t, user.ID, customer.ID, 200, 0, models.InvoiceStatusUnpaid)

	_, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 100,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv.ID, Amount: 50},
			{InvoiceID: inv.ID, Amount: 50},
		},
	}, user.ID)
	if err == nil || err.Error() != "isti račun ne može biti dodat dva puta u jednoj uplati" {
		t.Fatalf("expected duplicate, got %v", err)
	}
}

func TestCreatePayment_InactiveCustomerRejected(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Inactive", 100)
	database.DB.Model(&customer).Update("is_active", false)
	inv := seedOpenInvoice(t, user.ID, customer.ID, 100, 0, models.InvoiceStatusUnpaid)

	_, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 100,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv.ID, Amount: 100},
		},
	}, user.ID)
	if err == nil || err.Error() != "kupac nije aktivan" {
		t.Fatalf("expected inactive, got %v", err)
	}
}

func TestCreatePayment_RollbackOnError(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	customer := seedPaymentCustomer(t, "Rollback", 100)
	inv := seedOpenInvoice(t, user.ID, customer.ID, 100, 0, models.InvoiceStatusUnpaid)

	_, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID:  customer.ID,
		TotalAmount: 50,
		Allocations: []dto.CreatePaymentAllocationRequest{
			{InvoiceID: inv.ID, Amount: 30},
		},
	}, user.ID)
	if err == nil {
		t.Fatal("expected mismatch error")
	}

	var payCount int64
	database.DB.Model(&models.Payment{}).Count(&payCount)
	if payCount != 0 {
		t.Fatalf("payment rows after rollback=%d", payCount)
	}
	var updated models.Invoice
	database.DB.First(&updated, inv.ID)
	if updated.PaidAmount != 0 || updated.Status != models.InvoiceStatusUnpaid {
		t.Fatalf("invoice mutated on rollback: %+v", updated)
	}
	var c models.Customer
	database.DB.First(&c, customer.ID)
	if math.Abs(c.TotalDebt-100) > 0.01 {
		t.Fatalf("debt mutated=%v", c.TotalDebt)
	}
}

func TestGetAllPayments_FiltersAndPagination(t *testing.T) {
	setupPaymentTestDB(t)
	user := seedPaymentUser(t)
	c1 := seedPaymentCustomer(t, "A", 100)
	c2 := seedPaymentCustomer(t, "B", 200)
	inv1 := seedOpenInvoice(t, user.ID, c1.ID, 100, 0, models.InvoiceStatusUnpaid)
	inv2 := seedOpenInvoice(t, user.ID, c2.ID, 200, 0, models.InvoiceStatusUnpaid)

	_, err := CreatePayment(dto.CreatePaymentRequest{
		CustomerID: c1.ID, TotalAmount: 100,
		Allocations: []dto.CreatePaymentAllocationRequest{{InvoiceID: inv1.ID, Amount: 100}},
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = CreatePayment(dto.CreatePaymentRequest{
		CustomerID: c2.ID, TotalAmount: 200,
		Allocations: []dto.CreatePaymentAllocationRequest{{InvoiceID: inv2.ID, Amount: 200}},
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	list, total, err := GetAllPayments(PaymentListQuery{Page: 1, Limit: 10, CustomerID: ""})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("all total=%d len=%d", total, len(list))
	}

	filtered, fTotal, err := GetAllPayments(PaymentListQuery{
		Page: 1, Limit: 10, CustomerID: strconv.FormatUint(uint64(c1.ID), 10),
	})
	if err != nil {
		t.Fatal(err)
	}
	if fTotal != 1 || len(filtered) != 1 || filtered[0].CustomerID == nil || *filtered[0].CustomerID != c1.ID {
		t.Fatalf("filtered=%+v total=%d", filtered, fTotal)
	}

	page1, pTotal, err := GetAllPayments(PaymentListQuery{Page: 1, Limit: 1})
	if err != nil || pTotal != 2 || len(page1) != 1 {
		t.Fatalf("page1 len=%d total=%d err=%v", len(page1), pTotal, err)
	}

	loc, _ := time.LoadLocation("Europe/Belgrade")
	today := time.Now().In(loc).Truncate(24 * time.Hour)
	from := today
	toExclusive := today.AddDate(0, 0, 1)
	dated, dTotal, err := GetAllPayments(PaymentListQuery{
		Page: 1, Limit: 10, FromDate: &from, ToDate: &toExclusive,
	})
	if err != nil || dTotal < 1 || len(dated) < 1 {
		t.Fatalf("date filter total=%d len=%d err=%v", dTotal, len(dated), err)
	}
}
