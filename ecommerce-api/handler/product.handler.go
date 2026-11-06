package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"biteshare-api/ecommerce-api/model"
	"biteshare-api/ecommerce-api/repository"
	"biteshare-api/ecommerce-api/service"
	"biteshare-api/internal/auth"
	"biteshare-api/internal/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductHandler serves the public product catalog and admin product mutations.
type ProductHandler struct {
	repository *repository.ProductRepository
	uploader   *service.ImageKitUploader
}

func NewProductHandler(db database.Service, uploader *service.ImageKitUploader) *ProductHandler {
	return &ProductHandler{
		repository: repository.NewProductRepository(db),
		uploader:   uploader,
	}
}

// ListProducts retrieves all available products.
//
// @Summary Get all available products
// @Description Retrieve the public product catalog containing available food products.
// @Tags Products
// @Produce json
// @Success 200 {object} map[string]interface{} "List of available products"
// @Failure 500 {object} map[string]interface{} "Could not list products"
// @Router /ecom/products [get]
// ListProducts returns available products for the public catalog.
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var products []model.Product
	if err := h.repository.ListAvailable(&products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// GetProduct retrieves a product by its ID.
//
// @Summary Get product details
// @Description Retrieve details of one available product, including its images.
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{} "Product details"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Could not get product"
// @Router /ecom/product/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID, err := productIDFromPath(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var product model.Product
	if err := h.repository.FindAvailable(productID, &product); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

// CreateProduct creates a new food product.
//
// @Summary Create a product
// @Description Admin creates a product and uploads its image to ImageKit. Product information is submitted using multipart/form-data.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param product_name formData string true "Product name"
// @Param price formData number true "Product price"
// @Param stock formData number true "Available stock"
// @Param unit formData string true "Product unit"
// @Param description formData string false "Product description"
// @Param status formData string false "Product status"
// @Param expiry_date formData string false "Expiry date (YYYY-MM-DD)"
// @Param image formData file true "Product image"
// @Success 201 {object} map[string]interface{} "Product created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product information"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 502 {object} map[string]interface{} "Image upload failed"
// @Failure 503 {object} map[string]interface{} "ImageKit is not configured"
// @Failure 500 {object} map[string]interface{} "Could not create product"
// @Router /ecom/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin identity"})
		return
	}

	_, product, image, err := parseProductForm(c, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.uploader == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ImageKit is not configured"})
		return
	}

	imageURL, fileID, err := h.uploader.Upload(c.Request.Context(), image)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	product.AdminID = int32(adminID)
	productImage := model.ProductImage{ImageURL: imageURL, FileID: &fileID, IsPrimary: true}
	if err := h.repository.Create(int32(adminID), &product, &productImage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create product"})
		return
	}
	product.Images = []model.ProductImage{productImage}

	c.JSON(http.StatusCreated, gin.H{"product": product})
}

// UpdateProduct updates an existing product.
//
// @Summary Update a product
// @Description Admin updates product information. All fields are optional, but at least one field or a replacement image must be provided.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param product_name formData string false "Product name"
// @Param price formData number false "Product price"
// @Param stock formData number false "Available stock"
// @Param unit formData string false "Product unit"
// @Param description formData string false "Product description"
// @Param status formData string false "Product status"
// @Param expiry_date formData string false "Expiry date (YYYY-MM-DD)"
// @Param image formData file false "Replacement product image"
// @Success 200 {object} map[string]interface{} "Product updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product information"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 502 {object} map[string]interface{} "Image upload failed"
// @Failure 503 {object} map[string]interface{} "ImageKit is not configured"
// @Failure 500 {object} map[string]interface{} "Could not update product"
// @Router /ecom/products/{id} [patch]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin identity"})
		return
	}
	productID, err := productIDFromPath(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	values, _, image, err := parseProductForm(c, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(values) == 0 && image == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one product field or image is required"})
		return
	}

	var uploadedImage *model.ProductImage
	if image != nil {
		if h.uploader == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ImageKit is not configured"})
			return
		}
		imageURL, fileID, uploadErr := h.uploader.Upload(c.Request.Context(), image)
		if uploadErr != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": uploadErr.Error()})
			return
		}
		uploadedImage = &model.ProductImage{ImageURL: imageURL, FileID: &fileID, IsPrimary: true}
	}

	if err := h.repository.Update(int32(adminID), productID, values, uploadedImage); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}

// DeleteProduct deletes a product.
//
// @Summary Delete a product
// @Description Admin deletes a product and its associated images from ImageKit.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{} "Product deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 502 {object} map[string]interface{} "Failed to delete product image"
// @Failure 503 {object} map[string]interface{} "ImageKit is not configured"
// @Failure 500 {object} map[string]interface{} "Could not delete product"
// @Router /ecom/product/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	adminID, ok := auth.AdminID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin identity"})
		return
	}
	productID, err := productIDFromPath(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileIDs, err := h.repository.ImageFileIDs(int32(adminID), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find product images"})
		return
	}
	if len(fileIDs) > 0 && h.uploader == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ImageKit is not configured"})
		return
	}
	for _, fileID := range fileIDs {
		if err := h.uploader.Delete(c.Request.Context(), fileID); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
	}

	if err := h.repository.Delete(int32(adminID), productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}

func parseProductForm(c *gin.Context, required bool) (map[string]interface{}, model.Product, *multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		return nil, model.Product{}, nil, fmt.Errorf("request must be multipart/form-data")
	}

	values := make(map[string]interface{})
	product := model.Product{}
	fields := c.Request.MultipartForm.Value

	name, exists := fields["product_name"]
	if required && (!exists || strings.TrimSpace(first(name)) == "") {
		return nil, product, nil, fmt.Errorf("product_name is required")
	}
	if exists {
		product.Name = strings.TrimSpace(first(name))
		values["product_name"] = product.Name
	}

	if err := parseFloatField(fields, "price", required, &product.Price, values); err != nil {
		return nil, product, nil, err
	}
	if err := parseFloatField(fields, "stock", required, &product.Stock, values); err != nil {
		return nil, product, nil, err
	}

	unit, exists := fields["unit"]
	if required && (!exists || strings.TrimSpace(first(unit)) == "") {
		return nil, product, nil, fmt.Errorf("unit is required")
	}
	if exists {
		product.Unit = strings.TrimSpace(first(unit))
		values["unit"] = product.Unit
	}

	for _, field := range []string{"description", "status"} {
		if fieldValues, exists := fields[field]; exists {
			value := strings.TrimSpace(first(fieldValues))
			values[field] = value
			if field == "description" {
				product.Description = &value
			}
			if field == "status" {
				product.Status = value
			}
		}
	}
	if required && values["status"] == nil {
		values["status"] = "available"
		product.Status = "available"
	}

	if expiry, exists := fields["expiry_date"]; exists {
		expiryValue := strings.TrimSpace(first(expiry))
		if expiryValue == "" {
			values["expiry_date"] = nil
		} else {
			parsed, err := time.Parse("2006-01-02", expiryValue)
			if err != nil {
				return nil, product, nil, fmt.Errorf("expiry_date must use YYYY-MM-DD")
			}
			product.ExpiryDate = &parsed
			values["expiry_date"] = parsed
		}
	}

	image, err := c.FormFile("image")
	if required && err != nil {
		return nil, product, nil, fmt.Errorf("image is required")
	}
	if err != nil {
		image = nil
	}

	return values, product, image, nil
}

func parseFloatField(fields map[string][]string, name string, required bool, productValue *float64, values map[string]interface{}) error {
	field, exists := fields[name]
	if !exists && required {
		return fmt.Errorf("%s is required", name)
	}
	if !exists {
		return nil
	}

	parsed, err := strconv.ParseFloat(strings.TrimSpace(first(field)), 64)
	if err != nil || parsed < 0 {
		return fmt.Errorf("%s must be a non-negative number", name)
	}
	*productValue = parsed
	values[name] = parsed
	return nil
}

func productIDFromPath(c *gin.Context) (int32, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid product id")
	}
	return int32(id), nil
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
