package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestWebhook(t *testing.T) {
    handler := http.HandlerFunc(webhook)
    srv := httptest.NewServer(handler)
    defer  srv.Close()

    testCases := []struct {
        name         string
        method       string
        body         string
        expectedCode int
        expectedBody string
    }{
        {name: "method_get", method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
        {name: "method_put", method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
        {name: "method_delete", method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
        {name: "method_post_without_body", method: http.MethodPost, expectedCode: http.StatusInternalServerError, expectedBody: ""},
        {name: "method_post_unsupported_type", method: http.MethodPost, body:`{"request": {"type": "idunno", "command": "do something"}, "version": "1.0"}`, expectedCode: http.StatusUnprocessableEntity, expectedBody: ""},
        {
            name: "method_post_success",
            method: http.MethodPost,
            body: `{"request": {"type": "SimpleUtterance", "command": "sudo do something"}, "session": {"new": true}, "version": "1.0"}`,
            expectedCode: http.StatusOK,
            expectedBody: `Точное время .* часов, .* минут. Для вас нет новых сообщений.`,},
    }

    for _, tc := range testCases {
        t.Run(tc.method, func(t *testing.T) {
            // делаем запрос с помощью библиотеки resty к адресу запущенного сервера, 
            // который хранится в поле URL соответствующей структуры
            req := resty.New().R()
            req.Method = tc.method
            req.URL = srv.URL

            if len(tc.body) > 0 {
                req.SetHeader("Content-Type", "application/json")
                req.SetBody(tc.body)
            }

            resp, err := req.Send()
            assert.NoError(t, err, "error making HTTP request")

            assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")

            // проверим корректность полученного тела ответа, если мы его ожидаем
            if tc.expectedBody != "" {
                // assert.JSONEq помогает сравнить две JSON-строки
                assert.Regexp(t, tc.expectedBody, string(resp.Body()))
            }
        })
    }
}

