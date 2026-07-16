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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	databaseURL := "postgres://root:root123@localhost:5432/productdb_test?sslmode=disable"
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	_, _ = pool.Exec(ctx, "TRUNCATE TABLE products CASCADE")

	repo := database.NewProductRepositoryPGX(pool)

	p1 := domain.Product{Name: "測試商品 A", Description: "新增測試", Price: 100, Category: "test"}
	must(repo.Create(ctx, &p1))
	fmt.Printf("✅ 新增 — ID: %s, Name: %s\n", p1.ID, p1.Name)

	p2 := domain.Product{Name: "測試商品 B", Description: "修改前", Price: 200, Category: "test"}
	must(repo.Create(ctx, &p2))
	fmt.Printf("✅ 新增 — ID: %s, Name: %s\n", p2.ID, p2.Name)

	p2.Name = "測試商品 B (已修改)"
	p2.Description = "修改後"
	p2.Price = 250
	must(repo.Update(ctx, &p2))
	fmt.Printf("✅ 修改 — ID: %s, Name: %s\n", p2.ID, p2.Name)

	p3 := domain.Product{Name: "測試商品 C", Description: "將被刪除", Price: 300, Category: "test"}
	must(repo.Create(ctx, &p3))
	fmt.Printf("✅ 新增 — ID: %s, Name: %s\n", p3.ID, p3.Name)

	must(repo.Delete(ctx, p3.ID))
	fmt.Printf("✅ 刪除 — ID: %s\n\n", p3.ID)

	fmt.Println("=== 最終資料庫內容 ===")
	products, _ := repo.List(ctx)
	for _, p := range products {
		fmt.Printf("  • %s | %s | $%.2f | %s\n", p.ID, p.Name, p.Price, p.Description)
	}
	fmt.Printf("\n共 %d 筆資料\n", len(products))
}

func must(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
