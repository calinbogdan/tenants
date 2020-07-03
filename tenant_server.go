package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const contentType = "application/json"

type TenantServer struct {
	store TenantStore
	http.Handler
}

func NewTenantServer(store TenantStore) *TenantServer {
	server := new(TenantServer)

	router := mux.NewRouter()

	router.HandleFunc("/", server.getHandler).Methods("GET")
	router.HandleFunc("/", server.postHandler).Methods("POST")

	server.Handler = router
	server.store = store

	return server
}

func (t *TenantServer) getHandler(w http.ResponseWriter, r *http.Request) {
	userParams, exists := r.URL.Query()["user"]

	if !exists || len(userParams) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("[400]: URL 'user' parameter not specified or has specified too many.")
		return
	}

	w.Header().Set("content-type", contentType)

	userId := userParams[0]

	tenants, err := t.store.GetTenantsForUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print("[500]: Error at fetching tenants for user.")
		return
	}

	err = json.NewEncoder(w).Encode(tenants)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print("[500]: Error while trying to encode tenants slice to json.")
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

type CreateTenantRequestBody struct {
	TenantId string
	UserId   string
}

func (t *TenantServer) postHandler(w http.ResponseWriter, r *http.Request) {
	if requestHasNoBodyOrBodyIsNotJson(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body := new(CreateTenantRequestBody)

	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	tenants, err := t.store.GetTenantsForUser(body.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO this will be bad at some point -- once there are too many tenants, it will be a huge loop (pretty rare, but has to be fixed)
	for _, tenant := range tenants {
		if tenant.DatabaseId == body.TenantId {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	tenant, err := t.store.CreateTenant(body.TenantId, body.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tenant)
	return
}

func requestHasNoBodyOrBodyIsNotJson(r *http.Request) bool {
	return r.Body == http.NoBody || r.Body == nil || r.Header.Get("content-type") != contentType
}
