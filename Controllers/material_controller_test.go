package controllers

import (
	"Products/models"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
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

func TestGetMaterialByID(t *testing.T) {
	type testCase struct {
		name       string
		materialID string
		mockData   []interface{}
		expectErr  bool
		mockError  error
	}

	testCases := []testCase{
		{
			name:       "success - material found",
			materialID: "1",
			mockData: []interface{}{
				1, "Material 1", true, time.Now(), time.Now(), nil,
			},
			expectErr: false,
		},
		{
			name:       "material not found",
			materialID: "99",
			mockData:   nil,
			expectErr:  true,
		},
		{
			name:       "database error",
			materialID: "1",
			mockData:   nil,
			mockError:  errors.New("database error"),
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)

			query := regexp.QuoteMeta(`SELECT * FROM material WHERE id = $1 AND deleted_at IS NULL`)

			if tc.mockError != nil {
				mock.ExpectQuery(query).WithArgs(tc.materialID).WillReturnError(tc.mockError)
			} else if tc.mockData != nil {
				rowValues := make([]driver.Value, len(tc.mockData))
				for i, v := range tc.mockData {
					rowValues[i] = v
				}

				rows := sqlmock.NewRows([]string{"id", "name", "active", "created_at", "updated_at", "deleted_at"}).
					AddRow(rowValues...)

				mock.ExpectQuery(query).WithArgs(tc.materialID).WillReturnRows(rows).RowsWillBeClosed()
			}

			req := httptest.NewRequest("GET", "/materials/"+tc.materialID, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": tc.materialID})

			handler := GetMaterialByID(db)
			handler.ServeHTTP(w, req)

			// Debug response body
			fmt.Println("Response Code:", w.Code)
			fmt.Println("Response Body:", w.Body.String())

			if tc.expectErr {
				assert.Equal(t, http.StatusNotFound, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)

				var material models.Material
				err := json.NewDecoder(w.Body).Decode(&material)
				assert.NoError(t, err)
				assert.Equal(t, tc.mockData[1], material.Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateMaterial(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
		mockQueries  func()
	}{
		{
			name:         "success - valid request",
			requestBody:  `{"name": "Material 1", "active": true}`,
			expectedCode: http.StatusCreated,
			mockQueries: func() {
				mock.ExpectQuery(`INSERT INTO material \(name, active\) VALUES \(\$1, \$2\) RETURNING id, created_at, updated_at`).
					WithArgs("Material 1", true).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()))
			},
		},
		{
			name:         "failure - invalid JSON",
			requestBody:  `{"name": "Material 1", "active":}`,
			expectedCode: http.StatusBadRequest,
			mockQueries: func() {
				// No DB queries should run because JSON is invalid
			},
		},
		{
			name:         "failure - database error on insert",
			requestBody:  `{"name": "Material 1", "active": true}`,
			expectedCode: http.StatusInternalServerError,
			mockQueries: func() {
				mock.ExpectQuery(`INSERT INTO material \(name, active\) VALUES \(\$1, \$2\) RETURNING id, created_at, updated_at`).
					WithArgs("Material 1", true).
					WillReturnError(errors.New("insert error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockQueries()

			req := httptest.NewRequest("POST", "/materials", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := CreateMaterial(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateMaterial(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    testCases := []struct {
        name         string
        materialID   string
        requestBody  string
        expectedCode int
        mockQueries  func()
    }{
        {
            name:         "success - valid request",
            materialID:   "1",
            requestBody:  `{"name": "Updated Material", "active": true}`,
            expectedCode: http.StatusOK,
            mockQueries: func() {
                mock.ExpectExec(`UPDATE material SET name = \$1, active = \$2, updated_at = CURRENT_TIMESTAMP WHERE id = \$3 AND deleted_at IS NULL`).
                    WithArgs("Updated Material", true, "1").
                    WillReturnResult(sqlmock.NewResult(1, 1))
            },
        },
        {
            name:         "failure - invalid JSON",
            materialID:   "1",
            requestBody:  `{"name": "Updated Material", "active":}`,
            expectedCode: http.StatusBadRequest,
            mockQueries:  func() {},
        },
        {
            name:         "failure - database error on update",
            materialID:   "1",
            requestBody:  `{"name": "Updated Material", "active": true}`,
            expectedCode: http.StatusInternalServerError,
            mockQueries: func() {
                mock.ExpectExec(`UPDATE material SET name = \$1, active = \$2, updated_at = CURRENT_TIMESTAMP WHERE id = \$3 AND deleted_at IS NULL`).
                    WithArgs("Updated Material", true, "1").
                    WillReturnError(errors.New("update error"))
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            tc.mockQueries()

            req := httptest.NewRequest("PUT", "/materials/"+tc.materialID, strings.NewReader(tc.requestBody))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            req = mux.SetURLVars(req, map[string]string{"id": tc.materialID})

            handler := UpdateMaterial(db)
            handler.ServeHTTP(w, req)

            assert.Equal(t, tc.expectedCode, w.Code)
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}

func TestDeleteMaterial(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		materialID   string
		expectedCode int
		mockExec     func()
	}{
		{
			name:         "success - material deleted",
			materialID:   "1",
			expectedCode: http.StatusNoContent,
			mockExec: func() {
				mock.ExpectExec(`UPDATE material SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
			},
		},
		{
			name:         "failure - material not found",
			materialID:   "99",
			expectedCode: http.StatusNotFound,
			mockExec: func() {
				mock.ExpectExec(`UPDATE material SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(99).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
		},
		{
			name:         "failure - database error",
			materialID:   "1",
			expectedCode: http.StatusInternalServerError,
			mockExec: func() {
				mock.ExpectExec(`UPDATE material SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockExec()

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/material/%s", tc.materialID), nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.materialID})
			w := httptest.NewRecorder()

			handler := DeleteMaterial(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			err := mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

