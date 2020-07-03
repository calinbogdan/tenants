package main

import (
	"database/sql"
	"fmt"
	"log"
)

const getTenantsQuery = `SELECT Id, database_id FROM tenant 
	INNER JOIN tenant_member tm on tenant.Id = tm.tenant_id 
	WHERE tm.user_id = $1`

const insertTenants = `INSERT INTO tenant (database_id) VALUES ($1) RETURNING id`

const insertTenantMember = `INSERT INTO tenant_member (tenant_id, user_id) VALUES ($1, $2)`

const createDatabase = `CREATE DATABASE "%s"`

type PostgresTenantStore struct {
	tenantsUrl string
	tenantsServerUrl  string
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
	tenantsConnection, err := p.initTenantsConnection()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	tenantsServerConnection, err := p.initTenantsDatabaseServerConnection()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	tenantsConnectionTransaction, err := tenantsConnection.Begin()
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	// insert in tenant
	var newTenantId int
	err = tenantsConnectionTransaction.QueryRow(insertTenants, tenantId).Scan(&newTenantId)
	if err != nil {
		log.Fatalf("Insert tenant ERROR: %s", err.Error())
		return nil, err
	}

	// insert in tenant_member
	_, err = tenantsConnectionTransaction.Exec(insertTenantMember, newTenantId, userId)
	if err != nil {
		log.Fatalf("Insert team member ERROR: %s", err.Error())
		return nil, err
	}

	// create database
	createDatabaseQuery := fmt.Sprintf(createDatabase, tenantId)
	result, err := tenantsServerConnection.Exec(createDatabaseQuery)
	if err != nil {
		log.Print(result)
		tenantsConnectionTransaction.Rollback()
		log.Fatalf("Transaction rolled back, ERROR: %s", err.Error())
		return nil, err
	}
	tenantsConnectionTransaction.Commit()
	log.Print("Transaction committed")

	return nil, nil
}

func (p *PostgresTenantStore) initTenantsConnection() (*sql.DB, error) {
	return sql.Open("pgx", p.tenantsUrl)
}

func (p *PostgresTenantStore) initTenantsDatabaseServerConnection() (*sql.DB, error) {
	return sql.Open("pgx", p.tenantsServerUrl)
}
