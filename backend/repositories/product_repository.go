package repositories

import (
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	DefaultProductListPage  = 1
	DefaultProductListLimit = 20
	MaxProductListLimit     = 100
)

type ProductListQuery struct {
	Search          string
	CategoryID      string
	GroupID         string
	Ungrouped       bool
	IncludeInactive bool
	StockStatus     string
	Page            int
	Limit           int
}

func validateProductGroupAssignment(categoryID uint, groupID *uint) error {
	if groupID == nil {
		return nil
	}

	var group models.ProductGroup
	if err := database.DB.First(&group, *groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("grupa proizvoda nije pronađena")
		}
		return err
	}

	if group.CategoryID != categoryID {
		return errors.New("grupa ne pripada izabranoj kategoriji")
	}

	return nil
}

func buildProductListQuery(q ProductListQuery) *gorm.DB {
	query := database.DB.Model(&models.Product{})

	if !q.IncludeInactive {
		query = query.
			Where("products.is_active = ?", true).
			Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true)
	}

	if q.Search != "" {
		search := strings.ToLower(strings.TrimSpace(q.Search))
		pattern := "%" + search + "%"
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.slug) LIKE ?", pattern, pattern)
	}

	if q.CategoryID != "" {
		query = query.Where("products.category_id = ?", q.CategoryID)
	}
	if q.GroupID != "" {
		query = query.Where("products.group_id = ?", q.GroupID)
	}
	if q.Ungrouped {
		query = query.Where("products.group_id IS NULL")
	}

	switch strings.ToLower(strings.TrimSpace(q.StockStatus)) {
	case "out":
		query = query.Where("products.stock_quantity <= 0")
	}

	return query
}

func ListProducts(q ProductListQuery) ([]models.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = DefaultProductListPage
	}
	if q.Limit <= 0 {
		q.Limit = DefaultProductListLimit
	}

	var total int64
	if err := buildProductListQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []models.Product
	offset := (q.Page - 1) * q.Limit
	err := buildProductListQuery(q).
		Preload("Category").
		Preload("Group").
		Order("products.name ASC, products.id ASC").
		Limit(q.Limit).
		Offset(offset).
		Find(&products).Error
	return products, total, err
}

type PublicProductListQuery struct {
	Search        string
	CategoryID    string
	CategorySlug  string
	GroupID       string
	GroupSlug     string
	Ungrouped     bool
	OnSaleOnly    bool
	HomepageOnly  bool
	InStockOnly   bool
	ExcludeID     uint
	Random        bool
	Sort          string
	Page          int
	Limit         int
}

const (
	PublicSortRecommended = "recommended"
	PublicSortPriceAsc    = "price_asc"
	PublicSortPriceDesc   = "price_desc"
	PublicSortNameAsc     = "name_asc"
	PublicSortNameDesc    = "name_desc"
)

// Approximate effective price for server-side sort (discount applied; round-up to 10 not required for ordering).
const publicEffectivePriceExpr = `
CASE
  WHEN products.is_on_sale AND products.discount_percent > 0
  THEN products.sale_price * (1 - products.discount_percent / 100.0)
  ELSE products.sale_price
END
`

func NormalizePublicSort(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "", PublicSortRecommended, "default":
		return PublicSortRecommended
	case PublicSortPriceAsc, "price-asc", "price":
		return PublicSortPriceAsc
	case PublicSortPriceDesc, "price-desc":
		return PublicSortPriceDesc
	case PublicSortNameAsc, "name", "name-asc":
		return PublicSortNameAsc
	case PublicSortNameDesc, "name-desc":
		return PublicSortNameDesc
	default:
		return ""
	}
}

func publicProductOrderClause(sort string, random bool) string {
	if random {
		return "RANDOM()"
	}
	switch NormalizePublicSort(sort) {
	case PublicSortPriceAsc:
		return "(" + publicEffectivePriceExpr + ") ASC, products.name ASC, products.id ASC"
	case PublicSortPriceDesc:
		return "(" + publicEffectivePriceExpr + ") DESC, products.name ASC, products.id ASC"
	case PublicSortNameDesc:
		return "products.name DESC, products.id DESC"
	case PublicSortNameAsc:
		return "products.name ASC, products.id ASC"
	default:
		// recommended: featured first, then name
		return "products.show_on_homepage DESC, products.name ASC, products.id ASC"
	}
}

func buildPublicProductListQuery(q PublicProductListQuery) *gorm.DB {
	query := database.DB.Model(&models.Product{}).
		Where("products.is_active = ?", true).
		Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true)

	needsGroupJoin := q.GroupSlug != ""
	if needsGroupJoin {
		query = query.Joins("LEFT JOIN product_groups ON product_groups.id = products.group_id AND product_groups.deleted_at IS NULL")
	}

	if q.Search != "" {
		search := strings.ToLower(strings.TrimSpace(q.Search))
		pattern := "%" + search + "%"
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.slug) LIKE ?", pattern, pattern)
	}
	if q.CategorySlug != "" {
		query = query.Where("categories.slug = ?", strings.TrimSpace(q.CategorySlug))
	} else if q.CategoryID != "" {
		query = query.Where("products.category_id = ?", q.CategoryID)
	}
	if q.GroupSlug != "" {
		query = query.Where("product_groups.slug = ?", strings.TrimSpace(q.GroupSlug))
	} else if q.GroupID != "" {
		query = query.Where("products.group_id = ?", q.GroupID)
	}
	if q.Ungrouped {
		query = query.Where("products.group_id IS NULL")
	}
	if q.OnSaleOnly {
		query = query.Where("products.is_on_sale = ?", true)
	}
	if q.HomepageOnly {
		query = query.Where("products.show_on_homepage = ?", true)
	}
	if q.InStockOnly {
		query = query.Where("products.stock_quantity > ?", 0)
	}
	if q.ExcludeID > 0 {
		query = query.Where("products.id <> ?", q.ExcludeID)
	}
	return query
}

func ListPublicProducts(q PublicProductListQuery) ([]models.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = DefaultProductListPage
	}
	if q.Limit <= 0 {
		q.Limit = DefaultProductListLimit
	}
	if q.Random {
		q.Page = 1
	}

	var total int64
	if err := buildPublicProductListQuery(q).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []models.Product
	offset := (q.Page - 1) * q.Limit
	err := buildPublicProductListQuery(q).
		Preload("Category").
		Preload("Group").
		Order(publicProductOrderClause(q.Sort, q.Random)).
		Limit(q.Limit).
		Offset(offset).
		Find(&products).Error
	return products, total, err
}

func GetPublicCategoryBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := database.DB.
		Where("slug = ? AND is_active = ?", strings.TrimSpace(slug), true).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func GetPublicProductBySlug(slug string) (*models.Product, error) {
	var product models.Product
	result := database.DB.Model(&models.Product{}).
		Where("products.slug = ? AND products.is_active = ?", slug, true).
		Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true).
		Preload("Category").
		Preload("Group").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, sort_order ASC, id ASC")
		}).
		First(&product)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

// GetPublicProductByID returns an active product in an active category (public catalog rules).
func GetPublicProductByID(id uint) (*models.Product, error) {
	var product models.Product
	result := database.DB.Model(&models.Product{}).
		Where("products.id = ? AND products.is_active = ?", id, true).
		Joins("JOIN categories ON categories.id = products.category_id AND categories.deleted_at IS NULL AND categories.is_active = ?", true).
		Preload("Category").
		First(&product)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

func CreateProduct(product *models.Product) error {
	if err := validateCategoryActive(product.CategoryID); err != nil {
		return err
	}

	if err := validateProductGroupAssignment(product.CategoryID, product.GroupID); err != nil {
		return err
	}

	return database.DB.Create(product).Error
}

func GetProductById(id string) (*models.Product, error) {
	var product models.Product
	result := database.DB.
		Preload("Category").
		Preload("Group").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, sort_order ASC, id ASC")
		}).
		First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

func UpdateProduct(product *models.Product) error {
	if err := validateCategoryActive(product.CategoryID); err != nil {
		return err
	}

	if err := validateProductGroupAssignment(product.CategoryID, product.GroupID); err != nil {
		return err
	}

	// Explicit Select osigurava da se group_id može postaviti na NULL.
	return database.DB.Model(product).
		Select(
			"Name", "Slug", "Description", "CategoryID", "GroupID", "Unit",
			"SalePrice", "StockQuantity", "MinStockQuantity",
			"PurchasePrice", "MarginPercent", "VatPercent",
			"IsActive", "IsOnSale", "DiscountPercent", "ShowOnHomepage",
		).
		Updates(product).Error
}

func DeactivateProduct(id string) error {
	var productID uint
	if _, err := fmt.Sscanf(id, "%d", &productID); err != nil {
		return errors.New("proizvod nije pronađen")
	}

	hasImages, err := ProductHasImages(productID)
	if err != nil {
		return err
	}
	if hasImages {
		return ErrProductHasImages
	}

	result := database.DB.Model(&models.Product{}).Where("id = ?", id).Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("proizvod nije pronađen")
	}

	return nil
}

func ActivateProduct(id string) error {
	result := database.DB.Model(&models.Product{}).Where("id = ?", id).Update("is_active", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("proizvod nije pronađen")
	}
	return nil
}
