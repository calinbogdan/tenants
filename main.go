package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
	"os"
)

func GetTenantsHandler(w http.ResponseWriter, r *http.Request) {
	userParams, exists := r.URL.Query()["user"]

	if !exists || len(userParams) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tenantStore := PostgresTenantStore{"port=5432 host=localhost user=postgres password=postgres", "tenants"}

	userId := userParams[0]
	tenants, err := tenantStore.GetTenantsForUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenants)
	return
}

func PostTenantHandler(w http.ResponseWriter, r *http.Request) {
	var tenantMember CreateTenantRequestBody
	err := json.NewDecoder(r.Body).Decode(&tenantMember)

	if err != nil {
		return
	}

	tenantStore := PostgresTenantStore{"port=5432 host=localhost user=postgres password=postgres", "tenants"}
	tenantStore.CreateTenant(tenantMember.TenantId, tenantMember.UserId)
	w.WriteHeader(http.StatusCreated)
}

func main() {
	server := NewTenantServer(NewPostgresTenantStore(os.Getenv("TENANTS_URL"), os.Getenv("TENANTS_SERVER_URL")))

	fmt.Printf("TenantsUrl: %s", os.Getenv("TENANTS_URL"))
	http.ListenAndServe(":5000", server)
}
