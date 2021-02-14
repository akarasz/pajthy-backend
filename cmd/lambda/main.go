package main

import (
	"context"
	"math/rand"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/handler"
	"github.com/akarasz/pajthy-backend/store"
)

type Http struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type RequestContext struct {
	Http Http `json:"http"`
}

type Request struct {
	Headers        map[string]string `json:"headers"`
	QueryParams    map[string]string `json:"queryStringParameters"`
	RequestContext RequestContext    `json:"requestContext"`
	Body           string            `json:"body"`
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func HandleLambda(ctx context.Context, in Request) (Response, error) {
	req := httptest.NewRequest(
		in.RequestContext.Http.Method,
		in.RequestContext.Http.Path,
		strings.NewReader(in.Body))
	for k, v := range in.Headers {
		req.Header.Add(k, v)
	}
	q := req.URL.Query()
	for k, v := range in.QueryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	rr := httptest.NewRecorder()

	h := handler.New(store.NewInMemory(), event.New())
	h.ServeHTTP(rr, req)

	headers := map[string]string{}
	for k, vv := range rr.HeaderMap {
		headers[k] = vv[0]
	}

	return Response{
		StatusCode: rr.Code,
		Headers:    headers,
		Body:       rr.Body.String(),
	}, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	lambda.Start(HandleLambda)
}
