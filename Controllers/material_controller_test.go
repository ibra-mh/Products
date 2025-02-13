package controllers

import (
	"Products/models"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Implement Query method for MockDB
func TestGetMaterials(t *testing.T) {
	type testCase struct {
		name        string
		mockData    [][]interface{}
		expectedLen int
		mockError   error
	}

	testCases := []testCase{
		{
			name: "success - materials found",
			mockData: [][]interface{}{
				{1, "Material1", true, time.Now(), time.Now(), nil},
				{2, "Material2", false, time.Now(), time.Now(), nil},
			},
			expectedLen: 2,
		},
		{
			name:        "no materials found",
			mockData:    [][]interface{}{}, // No data
			expectedLen: 0,
		},
		{
			name:        "database error",
			mockData:    nil,
			expectedLen: 0,
			mockError:   errors.New("database error"),
		},
		{
			name:        "scan error",
			mockData:    [][]interface{}{{1, "Material1", "invalid_active", time.Now(), time.Now(), nil}},
			expectedLen: 0,
			mockError:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error initializing sqlmock: %v", err)
			}
			defer db.Close()

			// Define expected query behavior
			query := "SELECT * FROM material WHERE deleted_at IS NULL"

			if tc.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tc.mockError)
			} else {
				rows := sqlmock.NewRows([]string{"id", "name", "active", "created_at", "updated_at", "deleted_at"})
				for _, row := range tc.mockData {
					var values []driver.Value
					for _, v := range row {
						values = append(values, v)
					}
					rows.AddRow(values...)
				}
				mock.ExpectQuery(query).WillReturnRows(rows)
			}

			// Create test HTTP request
			req := httptest.NewRequest("GET", "/materials", nil)
			w := httptest.NewRecorder()

			// Call the handler
			handler := GetMaterials(db)
			handler.ServeHTTP(w, req)

			// Assert HTTP response
			if tc.mockError != nil || tc.name == "scan error" {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)

				// Decode response
				var materials []models.Material
				if err := json.NewDecoder(w.Body).Decode(&materials); err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				assert.Len(t, materials, tc.expectedLen)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
