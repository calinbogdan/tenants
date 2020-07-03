package main

type Tenant struct {
	Id         int    `json:"id"`
	DatabaseId string `json:"databaseId"`
}

type TenantMember struct {
	TenantId int
	UserId   string
}

type TenantStore interface {
	GetTenantsForUser(userId string) ([]Tenant, error)
	CreateTenant(tenantId, userId string) (*Tenant, error)
}
