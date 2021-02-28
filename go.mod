module github.com/akarasz/pajthy-backend

go 1.15

require (
	github.com/aws/aws-lambda-go v1.22.0
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.1.1
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.0.2
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression v1.0.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.1.1
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/stretchr/testify v1.6.1
	github.com/testcontainers/testcontainers-go v0.9.0
)

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69
