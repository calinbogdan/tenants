package main

import (
	bytes2 "bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTenantServer(t *testing.T) {
	server := NewTenantServer(NewStubTenantStore())

	t.Run("it should return 400 if 'user' query param is not provided", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, http.StatusBadRequest, response.Code)
	})

	t.Run("it should return 200 if 'user' query param is provided", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		q := request.URL.Query()
		q.Add("user", "randomUserId")

		request.URL.RawQuery = q.Encode()

		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, http.StatusOK, response.Code)
		assertContentType(t, contentType, response.Header().Get("content-type"))
	})
}

func TestPostTenantServerValidation(t *testing.T) {
	server := newTenantServer()

	t.Run("it should return 400 if content-type is not application/json", func(t *testing.T) {
		body := &CreateTenantRequestBody{"random-tenant", "random|userId"}
		bytes, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err.Error())
		}

		request := newPostTenantRequest(bytes)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, http.StatusBadRequest, response.Code)
	})

	t.Run("it should return 400 if 'userId' or 'tenantId' are not specified", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, http.StatusBadRequest, response.Code)
	})
}

func TestPostTenantServerToAvoidDuplicates(t *testing.T) {
	server := newTenantServer()

	t.Run("it should return 403 if there is a tenant with that id already", func(t *testing.T) {
		body := &CreateTenantRequestBody{"existing-tenant", "randomUserId"}
		bytes, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err.Error())
		}

		server.ServeHTTP(httptest.NewRecorder(), newPostTenantRequestWithJsonHeader(bytes))

		request := newPostTenantRequestWithJsonHeader(bytes)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, http.StatusForbidden, response.Code)
	})
}

func newTenantServer() *TenantServer {
	return NewTenantServer(NewStubTenantStore())
}

func newPostTenantRequest(bytes []byte) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, "/", bytes2.NewReader(bytes))
	return request
}

func newPostTenantRequestWithJsonHeader(bytes []byte) *http.Request {
	request := newPostTenantRequest(bytes)
	request.Header.Set("content-type", contentType)
	return request
}

func assertContentType(t *testing.T, expectedContentType string, actualContentType string) {
	t.Helper()
	if expectedContentType != actualContentType {
		t.Errorf("Expected content type '%s', got %s", expectedContentType, actualContentType)
	}
}

func assertStatusCode(t *testing.T, expectedCode int, actualCode int) {
	t.Helper()
	if expectedCode != actualCode {
		t.Errorf("Expected code %d, got code %d", expectedCode, actualCode)
	}
}
