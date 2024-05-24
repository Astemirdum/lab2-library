package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Astemirdum/library-service/backend/library/internal/handler"
	"github.com/Astemirdum/library-service/backend/library/internal/model"
	"github.com/Astemirdum/library-service/backend/pkg/validate"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	service_mocks "github.com/Astemirdum/library-service/backend/library/internal/handler/mocks"
)

func TestHandler_GetBooks(t *testing.T) {
	t.Parallel()
	type input struct {
		libraryUid string
		page, size int
		showAll    bool
	}
	type response struct {
		expectedCode int
		expectedBody string
	}
	type mockBehavior func(r *service_mocks.MockLibraryService, req input)

	var tests = []struct {
		name         string
		mockBehavior mockBehavior
		input        input
		response     response
		wantErr      bool
	}{
		{
			name: "ok",
			mockBehavior: func(r *service_mocks.MockLibraryService, req input) {
				r.EXPECT().
					ListBooks(context.Background(), req.libraryUid, req.showAll, req.page, req.size).
					Return(model.ListBooks{
						Paging: model.Paging{
							Page:          req.page,
							PageSize:      req.size,
							TotalElements: 1,
						},
						Items: []model.Book{
							{
								BookUid:        "f7cdc58f-2caf-4b15-9727-f89dcc629b27",
								Name:           "Краткий курс C++ в 7 томах",
								Author:         "Бьерн Страуструп",
								Genre:          "Научная фантастика",
								Condition:      "EXCELLENT",
								AvailableCount: 1,
							},
						},
					}, nil)
			},
			input: input{
				libraryUid: "83575e12-7ce0-48ee-9931-51919ff3c9ee",
				page:       0,
				size:       0,
				showAll:    false,
			},
			response: response{
				expectedCode: http.StatusOK,
				expectedBody: `{"page":0,"pageSize":0,"totalElements":1,"items":[{"id":0,"bookUid":"f7cdc58f-2caf-4b15-9727-f89dcc629b27","name":"Краткий курс C++ в 7 томах","author":"Бьерн Страуструп","genre":"Научная фантастика","condition":"EXCELLENT","availableCount":1}]}`,
			},
			wantErr: false,
		},
		{
			name:         "err. name required",
			mockBehavior: func(r *service_mocks.MockLibraryService, inp input) {},
			input: input{
				libraryUid: "",
				page:       0,
				size:       0,
				showAll:    false,
			},
			response: response{
				expectedCode: http.StatusBadRequest,
				expectedBody: `{"message":"empty libraryUid"}`,
			},
			wantErr: true,
		},
		{
			name: "err. internal",
			mockBehavior: func(r *service_mocks.MockLibraryService, inp input) {
				r.EXPECT().
					ListBooks(context.Background(), inp.libraryUid, inp.showAll, inp.page, inp.size).
					Return(model.ListBooks{}, errors.New("db internal"))
			},
			input: input{
				libraryUid: "83575e12-7ce0-48ee-9931-51919ff3c9ee",
				page:       0,
				size:       0,
				showAll:    false,
			},
			response: response{
				expectedCode: http.StatusInternalServerError,
				expectedBody: `{"message":"db internal"}`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := gomock.NewController(t)
			defer c.Finish()
			svc := service_mocks.NewMockLibraryService(c)
			log := zap.NewExample().Named("test")
			h := handler.New(svc, log)

			e := echo.New()
			e.Validator = validate.NewCustomValidator()
			e.GET("/libraries/:libraryUid/books", h.GetBooks)

			r := httptest.NewRequest(
				http.MethodGet, fmt.Sprintf("/libraries/%s/books?page=%d&size=%d&showAll=%v", tt.input.libraryUid, tt.input.page, tt.input.size, tt.input.showAll), http.NoBody)
			r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			tt.mockBehavior(svc, tt.input)
			e.ServeHTTP(w, r)

			require.Equal(t, tt.response.expectedCode, w.Code)
			require.Equal(t, tt.response.expectedBody, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
