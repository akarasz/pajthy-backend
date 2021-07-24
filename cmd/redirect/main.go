package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	PathParams map[string]string `json:"pathParameters"`
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}

func redirect(req *Request) (*Response, error) {
	return &Response{
		StatusCode: 301,
		Headers:    map[string]string{"Location": "https://5ulbtagtxf.execute-api.eu-central-1.amazonaws.com/test?session=" + req.PathParams["session"]},
	}, nil
}

func main() {
	lambda.Start(redirect)
}