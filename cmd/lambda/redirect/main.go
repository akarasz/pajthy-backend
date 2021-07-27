package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

const baseUrl = "https://5ulbtagtxf.execute-api.eu-central-1.amazonaws.com/test"

type Request struct {
	Headers    map[string]string `json:"headers"`
	PathParams map[string]string `json:"pathParameters"`
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}

func redirect(req *Request) (*Response, error) {
	return &Response{
		StatusCode: 301,
		Headers: map[string]string{
			"Location": fmt.Sprintf("%s?session=%s&type=%s",
				baseUrl,
				req.PathParams["session"],
				req.Headers["x-redirect-type"]),
		},
	}, nil
}

func main() {
	lambda.Start(redirect)
}
