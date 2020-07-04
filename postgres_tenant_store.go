package main

import (
	"database/sql"
	"fmt"
	"log"
)

const getTenantsQuery = `SELECT Id, database_id FROM tenant 
	INNER JOIN tenant_member tm on tenant.Id = tm.tenant_id 
	WHERE tm.user_id = $1`

const insertTenants = `INSERT INTO tenant (database_id) VALUES ($1) RETURNING *`

const insertTenantMember = `INSERT INTO tenant_member (tenant_id, user_id) VALUES ($1, $2)`

const createDatabase = `CREATE DATABASE "%s" WITH TEMPLATE template_tenantdb`

type PostgresTenantStore struct {
	tenantsUrl       string
	tenantsServerUrl string
}

func NewPostgresTenantStore(tenantsUrl, tenantsServerUrl string) *PostgresTenantStore {
	store := new(PostgresTenantStore)

	store.tenantsServerUrl = tenantsServerUrl
	store.tenantsUrl = tenantsUrl

	return store
}

func (p *PostgresTenantStore) GetTenantsForUser(userId string) (tenants []Tenant, err error) {
	connection, err := p.initTenantsConnection()
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	rows, err := connection.Query(getTenantsQuery, userId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var tenant Tenant
		if err := rows.Scan(&tenant.Id, &tenant.DatabaseId); err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

func (p *PostgresTenantStore) CreateTenant(tenantId string, userId string) (*Tenant, error) {
	// tenants table connection
	tenantsConnection, err := p.initTenantsConnection()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	defer tenantsConnection.Close()

	// tenants multi-tenant databases server connection
	tenantsServerConnection, err := p.initTenantsDatabaseServerConnection()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	defer tenantsServerConnection.Close()

	// transaction to create a tenant
	createTenantTransaction, err := tenantsConnection.Begin()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	// insert in tenant
	newTenant := new(Tenant)
	err = createTenantTransaction.QueryRow(insertTenants, tenantId).Scan(&newTenant.Id, &newTenant.DatabaseId)
	if err != nil {
		log.Fatalf("Insert tenant ERROR: %s", err.Error())
		return nil, err
	}

	// insert in tenant_member
	_, err = createTenantTransaction.Exec(insertTenantMember, newTenant.Id, userId)
	if err != nil {
		log.Fatalf("Insert team member ERROR: %s", err.Error())
		return nil, err
	}

	// prepare command and transaction to create database
	createDatabaseQuery := fmt.Sprintf(createDatabase, tenantId)

	// create database
	_, err = tenantsServerConnection.Exec(createDatabaseQuery)
	if err != nil {
		createTenantTransaction.Rollback()
		log.Fatalf("Transaction rolled back, ERROR: %s", err.Error())
		return nil, err
	}

	createTenantTransaction.Commit()
	log.Print("Transaction committed")
	return newTenant, nil
}

func (p *PostgresTenantStore) initTenantsConnection() (*sql.DB, error) {
	return sql.Open("pgx", p.tenantsUrl)
}

func (p *PostgresTenantStore) initTenantsDatabaseServerConnection() (*sql.DB, error) {
	return sql.Open("pgx", p.tenantsServerUrl)
}
