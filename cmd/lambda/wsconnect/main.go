package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type RequestContext struct {
	ConnectionID string `json:"connectionId"`
}

type Request struct {
	Context     RequestContext    `json:"requestContext"`
	QueryParams map[string]string `json:"queryStringParameters"`
}

type Response struct {
	StatusCode int `json:"statusCode"`
}

func redirect(req *Request) (*Response, error) {
	fmt.Println(req)
	fmt.Printf("connect to %s as %s, connectionID is %v",
		req.QueryParams["session"],
		req.QueryParams["type"],
		req.Context.ConnectionID)

	return &Response{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(redirect)
}
