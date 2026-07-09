package catalogrepository

import "github.com/afraniocaires/ecommerce/internal/catalog/domain"

func toProductModel(product *domain.Product) ProductModel {
	return ProductModel{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.PriceCents,
		Status:      string(product.Status),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func toProductEntity(productModel *ProductModel) (*domain.Product, error) {
	return domain.RestoreProduct(
		productModel.ID,
		productModel.Name,
		productModel.Description,
		productModel.PriceCents,
		domain.ProductStatus(productModel.Status),
		productModel.CreatedAt,
		productModel.UpdatedAt,
	)
}
