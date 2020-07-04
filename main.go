package main

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
	"os"
)

func main() {
	server := NewTenantServer(NewPostgresTenantStore(os.Getenv("TENANTS_URL"), os.Getenv("TENANTS_SERVER_URL")))

	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatal(err.Error())
	}
}
