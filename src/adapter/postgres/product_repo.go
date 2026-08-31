package postgres

import (
	"context"
	"errors"
	"fmt"

	domain "backend/src/domain/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewProductRepositoryPGX(pool *pgxpool.Pool) *ProductRepositoryPGX {
	return &ProductRepositoryPGX{pool: pool}
}

func (r *ProductRepositoryPGX) Create(
	ctx context.Context,
	product *domain.Product,
) error {
	status := product.Status
	if status == "" {
		status = "active"
	}

	memberID := product.CreatedBy
	if memberID == "" {
		memberID = "00000000-0000-0000-0000-000000000000"
	}

	var categoryID *string
	if product.CategoryID != "" {
		categoryID = &product.CategoryID
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product($1, $2, $3, $4)",
		product.Name, status, categoryID, memberID,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating product: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) List(
	ctx context.Context,
) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_products()")
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}

	products, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Product, error) {
		var p domain.Product

		err := row.Scan(
			&p.ID, &p.Name, &p.Status, &p.CategoryID, &p.Category,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return p, err
		}

		return p, nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}

	if products == nil {
		products = []domain.Product{}
	}

	return products, nil
}

func (r *ProductRepositoryPGX) Search(ctx context.Context, keyword, categoryID string) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT p.id, p.name, p.status, COALESCE(p.category_id::text, ''), COALESCE(c.name, ''), p.created_at, p.updated_at
		FROM products p LEFT JOIN categories c ON c.id = p.category_id
		WHERE ($1 = '' OR p.name ILIKE '%' || $1 || '%') AND (NULLIF($2, '') IS NULL OR p.category_id = NULLIF($2, '')::uuid)
		ORDER BY p.created_at DESC`, keyword, categoryID)
	if err != nil {
		return nil, fmt.Errorf("searching products: %w", err)
	}
	products, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Product, error) {
		var p domain.Product
		err := row.Scan(&p.ID, &p.Name, &p.Status, &p.CategoryID, &p.Category, &p.CreatedAt, &p.UpdatedAt)
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("searching products: %w", err)
	}
	if products == nil {
		products = []domain.Product{}
	}
	return products, nil
}

func (r *ProductRepositoryPGX) GetByID(
	ctx context.Context,
	id string,
) (*domain.Product, error) {
	var p domain.Product

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_by_id($1)", id).Scan(&p.ID, &p.Name, &p.Status, &p.CategoryID, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("getting product by ID: %w", err)
	}

	return &p, nil
}

func (r *ProductRepositoryPGX) Update(
	ctx context.Context,
	product *domain.Product,
) error {
	status := product.Status
	if status == "" {
		status = "active"
	}

	var categoryID *string
	if product.CategoryID != "" {
		categoryID = &product.CategoryID
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product($1, $2, $3, $4)",
		product.ID, product.Name, status, categoryID,
	).Scan(&product.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrProductNotFound
		}
		return fmt.Errorf("updating product: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) Delete(
	ctx context.Context,
	id string,
) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_product($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}
	if !deleted {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepositoryPGX) CreateDetail(
	ctx context.Context,
	detail *domain.ProductDetail,
) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_detail($1, $2, $3, $4)",
		detail.ProductID, detail.Introduction, detail.UsageInstructions, detail.ReturnPolicy,
	).Scan(&detail.ID, &detail.CreatedAt, &detail.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating detail: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) GetDetailByProductID(
	ctx context.Context,
	productID string,
) (*domain.ProductDetail, error) {
	var d domain.ProductDetail

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_detail_by_product($1)", productID).Scan(&d.ID, &d.ProductID, &d.Introduction, &d.UsageInstructions,
		&d.ReturnPolicy, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDetailNotFound
		}
		return nil, fmt.Errorf("getting detail by product ID: %w", err)
	}

	return &d, nil
}

func (r *ProductRepositoryPGX) UpdateDetail(
	ctx context.Context,
	detail *domain.ProductDetail,
) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product_detail($1, $2, $3, $4)",
		detail.ID, detail.Introduction, detail.UsageInstructions, detail.ReturnPolicy,
	).Scan(&detail.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrDetailNotFound
		}
		return fmt.Errorf("updating detail: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) GetPriceByID(
	ctx context.Context,
	id string,
) (*domain.ProductPrice, error) {
	var p domain.ProductPrice

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_price_by_id($1)", id).Scan(&p.ID, &p.ProductDetailID, &p.Label, &p.Amount, &p.Currency,
		&p.SortOrder, &p.ProductVariantID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("getting price by ID: %w", err)
	}

	return &p, nil
}

func (r *ProductRepositoryPGX) GetPricesByDetailID(
	ctx context.Context,
	detailID string,
) ([]domain.ProductPrice, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_product_prices_by_detail($1)", detailID)
	if err != nil {
		return nil, fmt.Errorf("listing prices by detail ID: %w", err)
	}

	prices, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ProductPrice, error) {
		var p domain.ProductPrice
		err := row.Scan(&p.ID, &p.ProductDetailID, &p.Label, &p.Amount,
			&p.Currency, &p.SortOrder, &p.ProductVariantID, &p.CreatedAt, &p.UpdatedAt)
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing prices by detail ID: %w", err)
	}

	if prices == nil {
		prices = []domain.ProductPrice{}
	}

	return prices, nil
}

func (r *ProductRepositoryPGX) UpdatePrice(
	ctx context.Context,
	price *domain.ProductPrice,
) error {
	var currency *string
	if price.Currency == "" {
		currency = nil
	} else {
		currency = &price.Currency
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product_price($1, $2, $3, $4, $5)",
		price.ID, price.Label, price.Amount, currency, price.SortOrder,
	).Scan(&price.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrPriceNotFound
		}
		return fmt.Errorf("updating price: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) CreatePrice(
	ctx context.Context,
	price *domain.ProductPrice,
) error {
	var currency *string
	if price.Currency == "" {
		currency = nil
	} else {
		currency = &price.Currency
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_price($1, $2, $3, $4, $5)",
		price.ProductDetailID, price.Label, price.Amount, currency, price.SortOrder,
	).Scan(&price.ID, &price.CreatedAt, &price.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating price: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) CreateOption(ctx context.Context, option *domain.ProductOption) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_option($1, $2, $3)", option.ProductDetailID, option.Name, option.Value).
		Scan(&option.ID, &option.ProductDetailID, &option.Name, &option.Value, &option.CreatedAt, &option.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating product option: %w", err)
	}
	return nil
}

func scanOption(row pgx.Row) (*domain.ProductOption, error) {
	var o domain.ProductOption
	err := row.Scan(&o.ID, &o.ProductDetailID, &o.Name, &o.Value, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}

func (r *ProductRepositoryPGX) GetOptionByID(ctx context.Context, id string) (*domain.ProductOption, error) {
	o, err := scanOption(r.pool.QueryRow(ctx, "SELECT * FROM get_product_option_by_id($1)", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductOptionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting product option: %w", err)
	}
	return o, nil
}

func (r *ProductRepositoryPGX) ListOptionsByDetailID(ctx context.Context, detailID string) ([]domain.ProductOption, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_product_options($1)", detailID)
	if err != nil {
		return nil, fmt.Errorf("listing product options: %w", err)
	}
	options, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ProductOption, error) {
		var o domain.ProductOption
		err := row.Scan(&o.ID, &o.ProductDetailID, &o.Name, &o.Value, &o.CreatedAt, &o.UpdatedAt)
		return o, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing product options: %w", err)
	}
	if options == nil {
		options = []domain.ProductOption{}
	}
	return options, nil
}

func (r *ProductRepositoryPGX) UpdateOption(ctx context.Context, option *domain.ProductOption) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product_option($1, $2, $3)", option.ID, option.Name, option.Value).Scan(&option.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrProductOptionNotFound
	}
	if err != nil {
		return fmt.Errorf("updating product option: %w", err)
	}
	return nil
}

func (r *ProductRepositoryPGX) DeleteOption(ctx context.Context, id string) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_product_option($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting product option: %w", err)
	}
	if !deleted {
		return domain.ErrProductOptionNotFound
	}
	return nil
}

func (r *ProductRepositoryPGX) CreateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_variant($1, $2, $3, $4, $5)", variant.ProductDetailID, variant.ProductPriceID, variant.SKU, variant.Status, variant.OptionIDs).
		Scan(&variant.ID, &variant.ProductDetailID, &variant.ProductPriceID, &variant.SKU, &variant.Status, &variant.CreatedAt, &variant.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "R0023" {
			return domain.ErrDuplicateSKU
		}
		return fmt.Errorf("creating product variant: %w", err)
	}
	return r.loadVariantOptions(ctx, variant)
}

func (r *ProductRepositoryPGX) loadVariantOptions(ctx context.Context, variant *domain.ProductVariant) error {
	return r.pool.QueryRow(ctx, "SELECT option_ids FROM get_product_variant_by_id($1)", variant.ID).Scan(&variant.OptionIDs)
}

func scanVariant(row pgx.Row) (*domain.ProductVariant, error) {
	var v domain.ProductVariant
	err := row.Scan(&v.ID, &v.ProductDetailID, &v.ProductPriceID, &v.SKU, &v.Status, &v.CreatedAt, &v.UpdatedAt, &v.OptionIDs)
	return &v, err
}

func (r *ProductRepositoryPGX) GetVariantByID(ctx context.Context, id string) (*domain.ProductVariant, error) {
	v, err := scanVariant(r.pool.QueryRow(ctx, "SELECT * FROM get_product_variant_by_id($1)", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductVariantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting product variant: %w", err)
	}
	return v, nil
}

func (r *ProductRepositoryPGX) ListVariantsByDetailID(ctx context.Context, detailID string) ([]domain.ProductVariant, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_product_variants($1)", detailID)
	if err != nil {
		return nil, fmt.Errorf("listing product variants: %w", err)
	}
	variants, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ProductVariant, error) {
		var v domain.ProductVariant
		err := row.Scan(&v.ID, &v.ProductDetailID, &v.ProductPriceID, &v.SKU, &v.Status, &v.CreatedAt, &v.UpdatedAt, &v.OptionIDs)
		return v, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing product variants: %w", err)
	}
	if variants == nil {
		variants = []domain.ProductVariant{}
	}
	return variants, nil
}

func (r *ProductRepositoryPGX) UpdateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product_variant($1, $2, $3, $4, $5)", variant.ID, variant.ProductPriceID, variant.SKU, variant.Status, variant.OptionIDs).Scan(&variant.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrProductVariantNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "R0023" {
			return domain.ErrDuplicateSKU
		}
		return fmt.Errorf("updating product variant: %w", err)
	}
	return r.loadVariantOptions(ctx, variant)
}

func (r *ProductRepositoryPGX) DeleteVariant(ctx context.Context, id string) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_product_variant($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting product variant: %w", err)
	}
	if !deleted {
		return domain.ErrProductVariantNotFound
	}
	return nil
}

func (r *ProductRepositoryPGX) CreateImage(ctx context.Context, image *domain.ProductImage) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_image($1, $2)", image.ProductID, image.Filename).Scan(&image.ID, &image.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating product image: %w", err)
	}
	return nil
}

func (r *ProductRepositoryPGX) ListImages(ctx context.Context, productID string) ([]domain.ProductImage, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_product_images($1)", productID)
	if err != nil {
		return nil, fmt.Errorf("listing product images: %w", err)
	}
	images, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ProductImage, error) {
		var image domain.ProductImage
		err := row.Scan(&image.ID, &image.ProductID, &image.Filename, &image.CreatedAt)
		return image, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing product images: %w", err)
	}
	if images == nil {
		images = []domain.ProductImage{}
	}
	return images, nil
}
