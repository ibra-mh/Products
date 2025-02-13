package controllers

import (
	"Products/models"
	"database/sql"
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
	"github.com/stretchr/testify/mock"
)

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type MockDB struct {
	mock.Mock
}

// Implement Query method for MockDB
func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	argsList := m.Called(append([]interface{}{query}, args...)...)
	if result, ok := argsList.Get(0).(*sql.Rows); ok {
		return result, argsList.Error(1)
	}
	return nil, argsList.Error(1)
}

func TestGetOffers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing sqlmock: %v", err)
	}
	defer db.Close()

	testCases := []struct {
		name         string
		mockData     [][]interface{}
		mockError    error
		expectedLen  int
		expectedCode int
	}{
		{
			name: "success - offers found",
			mockData: [][]interface{}{
				{1, "Offer1", time.Now(), time.Now(), nil},
				{2, "Offer2", time.Now(), time.Now(), nil},
			},
			expectedLen:  2,
			expectedCode: http.StatusOK,
		},
		{
			name:         "no offers found",
			mockData:     [][]interface{}{},
			expectedLen:  0,
			expectedCode: http.StatusOK,
		},
		{
			name:         "database error",
			mockData:     nil,
			mockError:    errors.New("database error"),
			expectedLen:  0,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := "SELECT * FROM offer WHERE deleted_at IS NULL"

			if tc.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tc.mockError)
			} else {
				rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "deleted_at"})
				for _, row := range tc.mockData {
					var values []driver.Value
					for _, v := range row {
						values = append(values, v)
					}
					rows.AddRow(values...)
				}
				mock.ExpectQuery(query).WillReturnRows(rows)
			}

			req := httptest.NewRequest("GET", "/offers", nil)
			w := httptest.NewRecorder()

			handler := GetOffers(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.mockError == nil {
				var offers []models.Offer
				if err := json.NewDecoder(w.Body).Decode(&offers); err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				assert.Len(t, offers, tc.expectedLen)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetOfferByID(t *testing.T) {
	type testCase struct {
		name      string
		offerID   string
		mockData  []interface{}
		expectErr bool
		mockError error
	}

	testCases := []testCase{
		{
			name:    "success - valid offer",
			offerID: "1",
			mockData: []interface{}{
				1, "Premium Plan", time.Now(), time.Now(), nil,
			},
			expectErr: false,
		},
		{
			name:      "offer not found",
			offerID:   "99",
			mockData:  nil,
			expectErr: true,
		},
		{
			name:      "database error",
			offerID:   "1",
			mockData:  nil,
			mockError: errors.New("database error"),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)

			query := regexp.QuoteMeta(`SELECT * FROM offer WHERE id = $1 AND deleted_at IS NULL`)

			if tc.mockError != nil {
				mock.ExpectQuery(query).WithArgs(tc.offerID).WillReturnError(tc.mockError)
			} else if tc.mockData != nil {
				rowValues := make([]driver.Value, len(tc.mockData))
				for i, v := range tc.mockData {
					rowValues[i] = v
				}

				rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "deleted_at"}).
					AddRow(rowValues...)

				mock.ExpectQuery(query).WithArgs(tc.offerID).WillReturnRows(rows).RowsWillBeClosed()
			}

			req := httptest.NewRequest("GET", "/offer/"+tc.offerID, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": tc.offerID})

			handler := GetOfferByID(db)
			handler.ServeHTTP(w, req)

			// Debugging logs
			fmt.Println("Response Code:", w.Code)
			fmt.Println("Response Body:", w.Body.String())

			if tc.expectErr {
				assert.Equal(t, http.StatusNotFound, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)

				var offer models.Offer
				err := json.NewDecoder(w.Body).Decode(&offer)
				assert.NoError(t, err)
				assert.Equal(t, tc.mockData[1], offer.Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateOffer(t *testing.T) {
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
			requestBody:  `{"name": "Premium Offer"}`,
			expectedCode: http.StatusCreated,
			mockQueries: func() {
				mock.ExpectQuery(`INSERT INTO offer \(name\) VALUES \(\$1\) RETURNING id, created_at, updated_at`).
					WithArgs("Premium Offer").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()))
			},
		},
		{
			name:         "failure - invalid JSON",
			requestBody:  `{"name": }`, // Malformed JSON
			expectedCode: http.StatusBadRequest,
			mockQueries: func() {
				// No DB queries should run because JSON is invalid
			},
		},
		{
			name:         "failure - database error on insert",
			requestBody:  `{"name": "Standard Offer"}`,
			expectedCode: http.StatusInternalServerError,
			mockQueries: func() {
				mock.ExpectQuery(`INSERT INTO offer \(name\) VALUES \(\$1\) RETURNING id, created_at, updated_at`).
					WithArgs("Standard Offer").
					WillReturnError(errors.New("insert error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockQueries()

			req := httptest.NewRequest("POST", "/offers", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := CreateOffer(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateOffer(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		offerID      string
		requestBody  string
		expectedCode int
		mockQueries  func()
	}{
		{
			name:         "success - valid request",
			offerID:      "1",
			requestBody:  `{"name": "Updated Offer Name"}`,
			expectedCode: http.StatusOK,
			mockQueries: func() {
				mock.ExpectExec(`UPDATE offer SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2 AND deleted_at IS NULL`).
					WithArgs("Updated Offer Name", "1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:         "failure - invalid JSON",
			offerID:      "1",
			requestBody:  `{"name": }`, // Malformed JSON
			expectedCode: http.StatusBadRequest,
			mockQueries:  func() {}, // No DB queries should run because JSON is invalid
		},
		{
			name:         "failure - database error on update",
			offerID:      "1",
			requestBody:  `{"name": "New Offer Name"}`,
			expectedCode: http.StatusInternalServerError,
			mockQueries: func() {
				mock.ExpectExec(`UPDATE offer SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2 AND deleted_at IS NULL`).
					WithArgs("New Offer Name", "1").
					WillReturnError(errors.New("update error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockQueries()

			req := httptest.NewRequest("PUT", "/offers/"+tc.offerID, strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": tc.offerID})

			handler := UpdateOffer(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteOffer(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		offerID      int
		expectedCode int
		mockExec     func()
	}{
		{
			name:         "success - offer deleted",
			offerID:      1,
			expectedCode: http.StatusNoContent,
			mockExec: func() {
				mock.ExpectExec(`UPDATE offer SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
			},
		},
		{
			name:         "failure - offer not found",
			offerID:      99,
			expectedCode: http.StatusNotFound,
			mockExec: func() {
				mock.ExpectExec(`UPDATE offer SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(99).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
		},
		{
			name:         "failure - database error",
			offerID:      1,
			expectedCode: http.StatusInternalServerError,
			mockExec: func() {
				mock.ExpectExec(`UPDATE offer SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockExec()

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/offer/%d", tc.offerID), nil)
			req = mux.SetURLVars(req, map[string]string{"id": fmt.Sprintf("%d", tc.offerID)})
			w := httptest.NewRecorder()

			handler := DeleteOffer(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			err := mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

