package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend/src/database"
	"backend/src/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	databaseURL := "postgres://root:root123@localhost:5432/productdb_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	_, _ = pool.Exec(ctx, "TRUNCATE TABLE inventory_items, inventories, product_prices, product_details, products, members CASCADE")

	repo := database.NewProductRepositoryPGX(pool)
	invRepo := database.NewInventoryRepositoryPGX(pool)

	// ── Product A: 演唱會門票 ──
	p1 := domain.Product{Name: "周杰倫演唱會門票", Price: 3800, Category: "ticket"}
	must(repo.Create(ctx, &p1))

	d1 := domain.ProductDetail{
		ProductID:         p1.ID,
		Introduction:      "2026 年全新巡迴演唱會",
		UsageInstructions: "憑 QR Code 入場",
		ReturnPolicy:      "恕不退換",
	}
	must(repo.CreateDetail(ctx, &d1))

	priceVIP := domain.ProductPrice{
		ProductDetailID: d1.ID,
		Label:           "VIP 區",
		Amount:          5800,
		Currency:        "TWD",
		SortOrder:       1,
	}
	must(repo.CreatePrice(ctx, &priceVIP))
	fmt.Printf("✅ 商品 A (演唱會) — 價格 ID: %s, %s $%.0f\n", priceVIP.ID, priceVIP.Label, priceVIP.Amount)

	priceStd := domain.ProductPrice{
		ProductDetailID: d1.ID,
		Label:           "一般區",
		Amount:          2800,
		Currency:        "TWD",
		SortOrder:       2,
	}
	must(repo.CreatePrice(ctx, &priceStd))
	fmt.Printf("✅ 商品 A (演唱會) — 價格 ID: %s, %s $%.0f\n", priceStd.ID, priceStd.Label, priceStd.Amount)

	// ── 庫存：VIP 區 ──
	invVIP := domain.Inventory{
		ProductPriceID: priceVIP.ID,
		Status:         "銷售中",
	}
	must(invRepo.CreateInventory(ctx, &invVIP))
	fmt.Printf("✅ 庫存總表 (VIP) — ID: %s\n", invVIP.ID)

	for i := 1; i <= 8; i++ {
		item := domain.InventoryItem{
			InventoryID: invVIP.ID,
			ItemCode:    fmt.Sprintf("VIP-%04d", i),
			Status:      "可用",
			Cost:        3000.00,
			DateAdded:   "2026-07-01",
		}
		must(invRepo.CreateItem(ctx, &item))
	}
	for i := 9; i <= 10; i++ {
		item := domain.InventoryItem{
			InventoryID: invVIP.ID,
			ItemCode:    fmt.Sprintf("VIP-%04d", i),
			Status:      "出售",
			Cost:        3000.00,
			DateAdded:   "2026-07-01",
		}
		must(invRepo.CreateItem(ctx, &item))
	}
	fmt.Printf("✅ 庫存明細 (VIP) — 10 筆 (可用 8, 出售 2)\n")

	// ── 庫存：一般區 ──
	invStd := domain.Inventory{
		ProductPriceID: priceStd.ID,
		Status:         "完售",
	}
	must(invRepo.CreateInventory(ctx, &invStd))

	for i := 1; i <= 5; i++ {
		item := domain.InventoryItem{
			InventoryID: invStd.ID,
			ItemCode:    fmt.Sprintf("STD-%04d", i),
			Status:      "出售",
			Cost:        1500.00,
			DateAdded:   "2026-07-01",
		}
		must(invRepo.CreateItem(ctx, &item))
	}
	fmt.Printf("✅ 庫存明細 (一般區) — 5 筆 (全部售出)\n")

	// ── Product B: 電子書 ──
	p2 := domain.Product{Name: "Go 語言實戰手冊", Price: 450, Category: "ebook"}
	must(repo.Create(ctx, &p2))

	d2 := domain.ProductDetail{
		ProductID:         p2.ID,
		Introduction:      "從入門到進階的 Go 語言學習指南",
		UsageInstructions: "購買後即可下載 PDF",
		ReturnPolicy:      "數位商品不適用七日鑑賞期",
	}
	must(repo.CreateDetail(ctx, &d2))

	priceEbook := domain.ProductPrice{
		ProductDetailID: d2.ID,
		Label:           "標準版",
		Amount:          450,
		Currency:        "TWD",
		SortOrder:       1,
	}
	must(repo.CreatePrice(ctx, &priceEbook))
	fmt.Printf("✅ 商品 B (電子書) — 價格 ID: %s, $%.0f\n", priceEbook.ID, priceEbook.Amount)

	priceBundle := domain.ProductPrice{
		ProductDetailID: d2.ID,
		Label:           "含原始碼套裝",
		Amount:          780,
		Currency:        "TWD",
		SortOrder:       2,
	}
	must(repo.CreatePrice(ctx, &priceBundle))
	fmt.Printf("✅ 商品 B (電子書) — 價格 ID: %s, $%.0f\n", priceBundle.ID, priceBundle.Amount)

	// ── 庫存：標準版 ──
	invEbook := domain.Inventory{
		ProductPriceID: priceEbook.ID,
		Status:         "銷售中",
	}
	must(invRepo.CreateInventory(ctx, &invEbook))
	fmt.Printf("✅ 庫存總表 (標準版) — ID: %s\n", invEbook.ID)

	for i := 1; i <= 4; i++ {
		item := domain.InventoryItem{
			InventoryID: invEbook.ID,
			ItemCode:    fmt.Sprintf("EB-STD-%04d", i),
			Status:      "可用",
			Cost:        200.00,
			DateAdded:   "2026-07-10",
		}
		must(invRepo.CreateItem(ctx, &item))
	}
	// 註銷 1 張
	must(invRepo.CreateItem(ctx, &domain.InventoryItem{
		InventoryID: invEbook.ID,
		ItemCode:    "EB-STD-0005",
		Status:      "註銷",
		Cost:        200.00,
		DateAdded:   "2026-07-10",
	}))
	// 賣出 2 張
	for i := 6; i <= 7; i++ {
		item := domain.InventoryItem{
			InventoryID: invEbook.ID,
			ItemCode:    fmt.Sprintf("EB-STD-%04d", i),
			Status:      "出售",
			Cost:        200.00,
			DateAdded:   "2026-07-10",
		}
		must(invRepo.CreateItem(ctx, &item))
	}
	fmt.Printf("✅ 庫存明細 (標準版) — 7 筆 (可用 4, 出售 2, 註銷 1)\n")

	// ── 庫存：含原始碼套裝（無庫存項目）──
	invBundle := domain.Inventory{
		ProductPriceID: priceBundle.ID,
		Status:         "註銷",
	}
	must(invRepo.CreateInventory(ctx, &invBundle))
	fmt.Printf("✅ 庫存總表 (套裝版) — ID: %s (已註銷，無明細)\n", invBundle.ID)

	fmt.Println("\n=== 資料庫範例資料已建立 ===")

	showProducts(ctx, pool)
}

func showProducts(ctx context.Context, pool *pgxpool.Pool) {
	rows, _ := pool.Query(ctx, `
		SELECT p.name, pr.label, pr.amount,
		       CONCAT(p.name, '-', pr.label) AS inventory_name,
		       i.status,
		       COUNT(it.id) AS total_quantity,
		       COUNT(it.id) FILTER (WHERE it.status = '出售') AS sold_quantity
		FROM products p
		JOIN product_details pd ON pd.product_id = p.id
		JOIN product_prices pr ON pr.product_detail_id = pd.id
		LEFT JOIN inventories i ON i.product_price_id = pr.id
		LEFT JOIN inventory_items it ON it.inventory_id = i.id
		GROUP BY p.name, pr.label, pr.amount, i.status, i.id
		ORDER BY p.name, pr.sort_order
	`)
	defer rows.Close()

	fmt.Println()
	for rows.Next() {
		var pName, prLabel, invName, iStatus string
		var amount float64
		var totalQty, soldQty int
		rows.Scan(&pName, &prLabel, &amount, &invName, &iStatus, &totalQty, &soldQty)
		fmt.Printf("  📦 %s -> %s | %s $%.0f [%s] %d/%d\n",
			pName, prLabel, invName, amount, iStatus, soldQty, totalQty)
	}
}

func must(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
