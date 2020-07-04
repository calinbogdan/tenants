package main

import (
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"net/http"
	"os"
)

func main() {
	server := NewTenantServer(NewPostgresTenantStore(os.Getenv("TENANTS_URL"), os.Getenv("TENANTS_SERVER_URL")))

	fmt.Printf("TenantsUrl: %s", os.Getenv("TENANTS_URL"))
	http.ListenAndServe(":5000", server)
}
