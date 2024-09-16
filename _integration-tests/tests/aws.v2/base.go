// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

//go:build integration

package awsv2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"orchestrion/integration/utils"
	"orchestrion/integration/validator/trace"
)

type base struct {
	server testcontainers.Container
	cfg    aws.Config
	host   string
	port   string
}

func (b *base) setup(t *testing.T) {
	b.server, b.host, b.port = utils.StartDynamoDBTestContainer(t)
}

func (b *base) teardown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, b.server.Terminate(ctx))
}

func (b *base) run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ddb := dynamodb.NewFromConfig(b.cfg)
	_, err := ddb.ListTables(ctx, nil)
	require.NoError(t, err)
}

func (b *base) expectedSpans() trace.Spans {
	return trace.Spans{
		{
			Tags: map[string]any{
				"name":     "DynamoDB.request",
				"service":  "aws.DynamoDB",
				"resource": "DynamoDB.ListTables",
				"type":     "http",
			},
			Meta: map[string]any{
				"aws.operation": "ListTables",
				"aws.region":    "test-region-1337",
				"aws_service":   "DynamoDB",
				"http.method":   "POST",
				"component":     "aws/aws-sdk-go-v2/aws",
				"span.kind":     "client",
			},
			Children: []*trace.Span{
				{
					Tags: map[string]any{
						"name":     "http.request",
						"service":  "aws.DynamoDB",
						"resource": "POST /",
						"type":     "http",
					},
					Meta: map[string]any{
						"http.method":              "POST",
						"http.status_code":         "200",
						"http.url":                 fmt.Sprintf("http://localhost:%s/", b.port),
						"network.destination.name": "localhost",
						"component":                "net/http",
						"span.kind":                "client",
					},
				},
			},
		},
	}
}