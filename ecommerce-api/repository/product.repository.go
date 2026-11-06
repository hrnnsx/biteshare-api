package repository

import (
	"biteshare-api/ecommerce-api/model"
	"biteshare-api/internal/database"
	"gorm.io/gorm"
)

// ProductRepository contains database operations for products.
type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db database.Service) *ProductRepository {
	return &ProductRepository{db: db.GORM()}
}

func (r *ProductRepository) ListAvailable(products *[]model.Product) error {
	return r.db.
		Preload("Images", imageOrder).
		Where("status = ?", "available").
		Order("product_id ASC").
		Find(products).Error
}

func (r *ProductRepository) FindAvailable(id int32, product *model.Product) error {
	return r.db.
		Preload("Images", imageOrder).
		Where("product_id = ? AND status = ?", id, "available").
		First(product).Error
}

func (r *ProductRepository) Create(adminID int32, product *model.Product, image *model.ProductImage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		product.AdminID = adminID
		// Save the image explicitly below instead of letting GORM auto-save
		// the Product.Images association as well.
		if err := tx.Omit("Images").Create(product).Error; err != nil {
			return err
		}

		if image != nil {
			image.ProductID = product.ID
			if err := tx.Create(image).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ProductRepository) Update(adminID, productID int32, values map[string]interface{}, image *model.ProductImage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Product{}).
			Where("product_id = ? AND admin_id = ?", productID, adminID).
			Updates(values)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if image != nil {
			if err := tx.Model(&model.ProductImage{}).
				Where("product_id = ?", productID).
				Update("is_primary", false).Error; err != nil {
				return err
			}
			image.ProductID = productID
			if err := tx.Create(image).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ProductRepository) ImageFileIDs(adminID, productID int32) ([]string, error) {
	var images []model.ProductImage
	result := r.db.
		Where("product_id = ? AND product_id IN (SELECT product_id FROM food_products WHERE product_id = ? AND admin_id = ?)", productID, productID, adminID).
		Find(&images)
	if result.Error != nil {
		return nil, result.Error
	}

	fileIDs := make([]string, 0, len(images))
	for _, image := range images {
		if image.FileID != nil && *image.FileID != "" {
			fileIDs = append(fileIDs, *image.FileID)
		}
	}
	return fileIDs, nil
}

func (r *ProductRepository) Delete(adminID, productID int32) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("product_id = ? AND admin_id = ?", productID, adminID).
			First(&model.Product{})
		if result.Error != nil {
			return result.Error
		}

		if err := tx.Where("product_id = ?", productID).Delete(&model.ProductImage{}).Error; err != nil {
			return err
		}

		return tx.Where("product_id = ? AND admin_id = ?", productID, adminID).
			Delete(&model.Product{}).Error
	})
}

func imageOrder(db *gorm.DB) *gorm.DB {
	return db.Order("is_primary DESC, image_id ASC")
}
