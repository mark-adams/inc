package main

import (
	"log"
	"os"

	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/mark-adams/inc/backends"
)

const testDbStr = "memory://"

func resetDatabase(t *testing.T) {
	os.Setenv("DB_URL", testDbStr)
	db, err := backends.GetBackend()
	if err != nil {
		t.Fatalf("Error connecting to database: %s", err)
	}

	// Drop tables
	if err = db.DropSchema(); err != nil {
		t.Fatalf("Error dropping schema: %s", err)
	}

	// Create tables
	if err = db.CreateSchema(); err != nil {
		t.Fatalf("Error creating tables: %s", err)
	}
}

func TestNewTokenCreationReturnsResponse(t *testing.T) {
	resetDatabase(t)
	req, err := http.NewRequest("POST", "http://localhost/new", nil)
	if err != nil {
		log.Fatal(err)
	}

	response := httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 201 {
		t.Fatalf("Incorrect status code: expected %v, actual: %v", 201, response.Code)
	}

	body := response.Body.String()
	if len(body) != 32 {
		t.Fatalf("Incorrect response length: expected %v, actual: %v", 32, len(body))
	}

}

func TestNewTokensAreUnique(t *testing.T) {
	resetDatabase(t)

	req, err := http.NewRequest("POST", "http://localhost/new", nil)
	if err != nil {
		log.Fatal(err)
	}

	firstResponse := httptest.NewRecorder()
	app.ServeHTTP(firstResponse, req)

	secondResponse := httptest.NewRecorder()
	app.ServeHTTP(secondResponse, req)

	if firstResponse.Body.Len() != secondResponse.Body.Len() {
		t.Fatalf(
			"Response lengths should be the same: first: %v, second: %v",
			firstResponse.Body.Len(),
			secondResponse.Body.Len(),
		)
	}

	if firstResponse.Body.String() == secondResponse.Body.String() {
		t.Fatalf(
			"Tokens should not match: %v == %v",
			firstResponse.Body.String(),
			secondResponse.Body.String(),
		)
	}
}

func TestPutOnNewTokenShouldIncrementValue(t *testing.T) {
	resetDatabase(t)
	req, err := http.NewRequest("POST", "http://localhost/new", nil)
	if err != nil {
		log.Fatal(err)
	}

	response := httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 201 {
		t.Fatalf("Incorrect status code: expected %v, actual %v", 201, response.Code)
	}

	token := response.Body.String()
	req, err = http.NewRequest("PUT", "http://localhost/"+token, nil)
	if err != nil {
		log.Fatal(err)
	}

	response = httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 200 {
		t.Fatalf("Incorrect status code: expected %v, actual %v", 200, response.Code)
	}

	if response.Body.String() != "1" {
		t.Fatalf("Incorrect initial token value: expected %v, actual %v", "1", response.Body.String())
	}

	response = httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 200 {
		t.Fatalf("Incorrect status code: expected %v, actual %v", 200, response.Code)
	}

	if response.Body.String() != "2" {
		t.Fatalf("Counter value did not increment: expected %v, actual %v", "2", response.Body.String())
	}

}

func TestPutOnNewNamespaceShouldIncrementFromZero(t *testing.T) {
	resetDatabase(t)
	req, err := http.NewRequest("POST", "http://localhost/new", nil)
	if err != nil {
		log.Fatal(err)
	}

	response := httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 201 {
		t.Fatalf("Incorrect status code: expected %v, actual %v", 201, response.Code)
	}

	token := response.Body.String()
	req, err = http.NewRequest("PUT", "http://localhost/"+token+"/1.0", nil)
	if err != nil {
		log.Fatal(err)
	}

	response = httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 200 {
		t.Fatalf("Incorrect status code: expected %v, actual %v", 200, response.Code)
	}

	if response.Body.String() != "0" {
		t.Fatalf("Incorrect initial token value: expected %v, actual %v", "0", response.Body.String())
	}

	response = httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 200 {
		t.Fatalf("Incorrect status code: expected %v, actual %v", 200, response.Code)
	}

	if response.Body.String() != "1" {
		t.Fatalf("Counter value did not increment: expected %v, actual %v", "1", response.Body.String())
	}

	// Change namespace from 1.0 -> 1.1
	req, err = http.NewRequest("PUT", "http://localhost/"+token+"/1.1", nil)
	if err != nil {
		log.Fatal(err)
	}

	response = httptest.NewRecorder()
	app.ServeHTTP(response, req)

	if response.Code != 200 {
		t.Fatalf("Incorrect status code: expected %v, actual %v", 200, response.Code)
	}

	if response.Body.String() != "0" {
		t.Fatalf("Incorrect initial token value for new namespace: expected %v, actual %v", "0", response.Body.String())
	}

}
