package main

type StubTenantStore struct {
	tenants        map[int]Tenant
	tenantsMembers []TenantMember
}

func NewStubTenantStore() *StubTenantStore {
	store := new (StubTenantStore)

	store.tenants = make(map[int]Tenant)
	return store
}

func (i *StubTenantStore) GetTenantsForUser(userId string) ([]Tenant, error) {
	var userTenants []Tenant
	for _, tenantMember := range i.tenantsMembers {
		if tenantMember.UserId == userId {
			userTenants = append(userTenants, i.tenants[tenantMember.TenantId])
		}
	}
	return userTenants, nil
}

func (i *StubTenantStore) CreateTenant(tenantId, userId string) (*Tenant, error) {
	nextId := i.getHighestId() + 1

	newTenant := Tenant{nextId, tenantId}
	i.tenants[nextId] = newTenant
	i.tenantsMembers = append(i.tenantsMembers, TenantMember{nextId, userId})

	return &newTenant, nil
}

func (i *StubTenantStore) getHighestId() int {
	id := 0
	for key, _ := range i.tenants {
		if key > id {
			id = key
		}
	}

	return id
}


