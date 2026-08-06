package invoicepdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"strings"

	"github.com/go-pdf/fpdf"
)

//go:embed assets/DejaVuSans.ttf
var fontRegular []byte

//go:embed assets/DejaVuSans-Bold.ttf
var fontBold []byte

//go:embed assets/logo-stampa-racuni.jpeg
var logoJPEG []byte

// Company holds optional firm details for the PDF header.
// Empty fields are omitted from the document.
type Company struct {
	Name               string
	Address            string
	City               string
	Phone              string
	Email              string
	TaxID              string
	RegistrationNumber string
	BankAccount        string
	Website            string
}

// Item is one invoice line for PDF rendering.
type Item struct {
	ProductName string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	TotalPrice  float64
}

// Document is the invoice data needed to build an A4 PDF.
type Document struct {
	ID              uint
	CreatedAt       string
	Status          string
	CustomerName    string
	CustomerPhone   string
	IsCashSale      bool
	TotalAmount     float64
	PaidAmount      float64
	RemainingAmount float64
	CreatedBy       string
	Items           []Item
	Company         Company
}

func (c Company) displayName() string {
	if strings.TrimSpace(c.Name) == "" {
		return "AM Keramika"
	}
	return strings.TrimSpace(c.Name)
}

func statusLabel(status string) string {
	switch status {
	case "paid":
		return "Plaćeno"
	case "unpaid":
		return "Neplaćeno"
	case "partially_paid":
		return "Djelimično plaćeno"
	case "cancelled":
		return "Otkazano"
	default:
		return status
	}
}

func formatMoney(amount float64) string {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		amount = 0
	}
	neg := amount < 0
	if neg {
		amount = -amount
	}
	cents := int64(math.Round(amount * 100))
	whole := cents / 100
	frac := cents % 100
	wholeStr := formatThousands(whole)
	var out string
	if frac == 0 {
		out = wholeStr + " RSD"
	} else {
		out = fmt.Sprintf("%s,%02d RSD", wholeStr, frac)
	}
	if neg {
		return "-" + out
	}
	return out
}

func formatThousands(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	rem := len(s) % 3
	if rem == 0 {
		rem = 3
	}
	b.WriteString(s[:rem])
	for i := rem; i < len(s); i += 3 {
		b.WriteByte('.')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func formatQuantity(q float64) string {
	if math.IsNaN(q) || math.IsInf(q, 0) {
		return "0"
	}
	// Trim trailing zeros but keep up to 3 decimals for m2/kg/l.
	s := fmt.Sprintf("%.3f", q)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	s = strings.ReplaceAll(s, ".", ",")
	return s
}

// Generate builds an A4 portrait PDF for the invoice document.
func Generate(doc Document) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AddUTF8FontFromBytes("dejavu", "", fontRegular)
	pdf.AddUTF8FontFromBytes("dejavu", "B", fontBold)
	pdf.AddPage()

	drawHeader(pdf, doc)
	drawCustomer(pdf, doc)
	drawItemsTable(pdf, doc)
	drawTotals(pdf, doc)
	drawFooter(pdf, doc)

	if doc.Status == "cancelled" {
		drawCancelledWatermark(pdf)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawHeader(pdf *fpdf.Fpdf, doc Document) {
	company := doc.Company
	name := company.displayName()

	logoH := 18.0
	logoW := 0.0
	if len(logoJPEG) > 0 {
		opt := fpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}
		info := pdf.RegisterImageOptionsReader("logo", opt, bytes.NewReader(logoJPEG))
		if info != nil {
			logoW = logoH * info.Width() / info.Height()
			if logoW > 55 {
				logoW = 55
				logoH = logoW * info.Height() / info.Width()
			}
			pdf.ImageOptions("logo", 14, 14, logoW, logoH, false, opt, 0, "")
		}
	}

	leftX := 14.0
	textX := leftX
	if logoW > 0 {
		textX = leftX + logoW + 4
	}
	y := 14.0
	pdf.SetXY(textX, y)
	pdf.SetFont("dejavu", "B", 12)
	pdf.CellFormat(90, 6, name, "", 1, "L", false, 0, "")

	pdf.SetFont("dejavu", "", 8)
	pdf.SetTextColor(80, 80, 80)
	writeOptionalLine(pdf, textX, company.Address)
	writeOptionalLine(pdf, textX, company.City)
	if phone := strings.TrimSpace(company.Phone); phone != "" {
		writeOptionalLine(pdf, textX, "Tel: "+phone)
	}
	writeOptionalLine(pdf, textX, company.Email)
	writeOptionalLine(pdf, textX, company.Website)
	if tax := strings.TrimSpace(company.TaxID); tax != "" {
		writeOptionalLine(pdf, textX, "PIB: "+tax)
	}
	if mb := strings.TrimSpace(company.RegistrationNumber); mb != "" {
		writeOptionalLine(pdf, textX, "MB: "+mb)
	}
	if bank := strings.TrimSpace(company.BankAccount); bank != "" {
		writeOptionalLine(pdf, textX, "Žiro račun: "+bank)
	}
	pdf.SetTextColor(0, 0, 0)

	pdf.SetXY(120, 14)
	pdf.SetFont("dejavu", "B", 18)
	pdf.CellFormat(76, 8, "RAČUN", "", 1, "R", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("dejavu", "B", 11)
	pdf.CellFormat(76, 6, fmt.Sprintf("Br. %d", doc.ID), "", 1, "R", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("dejavu", "", 9)
	pdf.CellFormat(76, 5, doc.CreatedAt, "", 1, "R", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("dejavu", "B", 9)
	pdf.CellFormat(76, 5, statusLabel(doc.Status), "", 1, "R", false, 0, "")

	headerBottom := pdf.GetY()
	if headerBottom < 14+logoH+2 {
		headerBottom = 14 + logoH + 2
	}
	pdf.SetY(headerBottom + 4)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(14, pdf.GetY(), 196, pdf.GetY())
	pdf.Ln(4)
}

func writeOptionalLine(pdf *fpdf.Fpdf, x float64, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	pdf.SetX(x)
	pdf.CellFormat(90, 4, value, "", 1, "L", false, 0, "")
}

func drawCustomer(pdf *fpdf.Fpdf, doc Document) {
	pdf.SetFont("dejavu", "B", 8)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 5, "KUPAC", "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("dejavu", "B", 10)
	if doc.IsCashSale {
		pdf.CellFormat(0, 5, "Gotovinska prodaja", "", 1, "L", false, 0, "")
	} else {
		name := strings.TrimSpace(doc.CustomerName)
		if name == "" {
			name = "Kupac"
		}
		pdf.CellFormat(0, 5, name, "", 1, "L", false, 0, "")
		if phone := strings.TrimSpace(doc.CustomerPhone); phone != "" {
			pdf.SetFont("dejavu", "", 9)
			pdf.CellFormat(0, 4.5, phone, "", 1, "L", false, 0, "")
		}
	}
	pdf.Ln(3)
}

func drawItemsTable(pdf *fpdf.Fpdf, doc Document) {
	headers := []string{"R.br.", "Proizvod", "Količina", "Jed.", "Jed. cijena", "Ukupno"}
	widths := []float64{12, 72, 22, 16, 30, 30}

	pdf.SetFont("dejavu", "B", 8)
	pdf.SetFillColor(245, 245, 244)
	pdf.SetDrawColor(160, 160, 160)
	for i, h := range headers {
		align := "L"
		if i >= 2 {
			align = "R"
		}
		if i == 3 {
			align = "L"
		}
		pdf.CellFormat(widths[i], 7, h, "B", 0, align, true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("dejavu", "", 8)
	for i, item := range doc.Items {
		ensureSpace(pdf, 8)
		name := strings.TrimSpace(item.ProductName)
		if name == "" {
			name = "Proizvod"
		}
		unit := strings.TrimSpace(item.Unit)
		if unit == "" {
			unit = "—"
		}
		row := []string{
			fmt.Sprintf("%d", i+1),
			name,
			formatQuantity(item.Quantity),
			unit,
			formatMoney(item.UnitPrice),
			formatMoney(item.TotalPrice),
		}
		// Wrap long product names by estimating height.
		lines := pdf.SplitLines([]byte(name), widths[1]-1)
		rowH := 6.0
		if len(lines) > 1 {
			rowH = float64(len(lines)) * 4.2
			if rowH < 6 {
				rowH = 6
			}
		}
		ensureSpace(pdf, rowH+1)
		y := pdf.GetY()
		x := 14.0
		for col, text := range row {
			align := "L"
			if col == 2 || col == 4 || col == 5 {
				align = "R"
			}
			if col == 1 && len(lines) > 1 {
				pdf.SetXY(x, y)
				pdf.MultiCell(widths[col], 4.2, text, "", "L", false)
				pdf.SetXY(x+widths[col], y)
			} else {
				pdf.SetXY(x, y)
				pdf.CellFormat(widths[col], rowH, text, "", 0, align, false, 0, "")
			}
			x += widths[col]
		}
		pdf.SetY(y + rowH)
		pdf.SetDrawColor(220, 220, 220)
		pdf.Line(14, pdf.GetY(), 196, pdf.GetY())
	}

	if len(doc.Items) == 0 {
		pdf.SetFont("dejavu", "", 9)
		pdf.CellFormat(0, 8, "Nema stavki.", "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)
}

func ensureSpace(pdf *fpdf.Fpdf, needed float64) {
	_, pageH := pdf.GetPageSize()
	_, _, _, bottom := pdf.GetMargins()
	if pdf.GetY()+needed > pageH-bottom {
		pdf.AddPage()
	}
}

func drawTotals(pdf *fpdf.Fpdf, doc Document) {
	ensureSpace(pdf, 28)
	pdf.SetX(120)
	pdf.SetFont("dejavu", "", 9)
	writeTotalRow(pdf, "Ukupno", formatMoney(doc.TotalAmount), false)
	writeTotalRow(pdf, "Plaćeno", formatMoney(doc.PaidAmount), false)

	label := "Preostalo"
	value := formatMoney(doc.RemainingAmount)
	if doc.Status == "cancelled" {
		label = "Status"
		value = statusLabel("cancelled")
	} else if doc.RemainingAmount <= 0.000001 {
		label = "Status plaćanja"
		value = "Plaćeno"
	}
	writeTotalRow(pdf, label, value, true)
	pdf.Ln(6)
}

func writeTotalRow(pdf *fpdf.Fpdf, label, value string, bold bool) {
	pdf.SetX(120)
	if bold {
		pdf.SetFont("dejavu", "B", 10)
	} else {
		pdf.SetFont("dejavu", "", 9)
	}
	pdf.CellFormat(36, 6, label, "", 0, "L", false, 0, "")
	pdf.CellFormat(40, 6, value, "", 1, "R", false, 0, "")
}

func drawFooter(pdf *fpdf.Fpdf, doc Document) {
	ensureSpace(pdf, 40)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(14, pdf.GetY(), 196, pdf.GetY())
	pdf.Ln(5)

	pdf.SetFont("dejavu", "", 8)
	pdf.SetTextColor(80, 80, 80)
	if createdBy := strings.TrimSpace(doc.CreatedBy); createdBy != "" {
		pdf.CellFormat(0, 4.5, "Račun kreirao: "+createdBy, "", 1, "L", false, 0, "")
	}
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)

	y := pdf.GetY()
	pdf.Line(14, y+12, 80, y+12)
	pdf.SetXY(14, y+13)
	pdf.SetFont("dejavu", "", 8)
	pdf.CellFormat(66, 4, "Potpis kupca", "", 0, "L", false, 0, "")

	pdf.Line(130, y+12, 196, y+12)
	pdf.SetXY(130, y+13)
	pdf.CellFormat(66, 4, "Potpis / pečat prodavca", "", 0, "R", false, 0, "")
}

func drawCancelledWatermark(pdf *fpdf.Fpdf) {
	pageCount := pdf.PageCount()
	for i := 1; i <= pageCount; i++ {
		pdf.SetPage(i)
		pdf.SetTextColor(180, 180, 180)
		pdf.SetFont("dejavu", "B", 54)
		pdf.TransformBegin()
		pdf.TransformRotate(35, 105, 150)
		pdf.SetXY(40, 145)
		pdf.CellFormat(140, 20, "OTKAZANO", "", 0, "C", false, 0, "")
		pdf.TransformEnd()
		pdf.SetTextColor(0, 0, 0)
	}
}

// Filename returns the standard download/inline PDF name.
func Filename(id uint) string {
	return fmt.Sprintf("AM-Keramika-Racun-%d.pdf", id)
}
