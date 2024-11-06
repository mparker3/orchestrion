// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

//go:build integration

package logrus

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"datadoghq.dev/orchestrion/_integration-tests/validator/trace"
)

type TestCaseStructLiteralPtr struct {
	logger *logrus.Logger
	logs   *bytes.Buffer
}

func (tc *TestCaseStructLiteralPtr) Setup(*testing.T) {
	tc.logs = new(bytes.Buffer)
	tc.logger = &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	tc.logger.SetLevel(logrus.DebugLevel)
	tc.logger.SetOutput(tc.logs)
}

func (tc *TestCaseStructLiteralPtr) Run(t *testing.T) {
	runTest(t, tc.logs, tc.Log)
}

func (*TestCaseStructLiteralPtr) Teardown(*testing.T) {}

func (*TestCaseStructLiteralPtr) ExpectedTraces() trace.Traces {
	return expectedTraces()
}

//dd:span
func (tc *TestCaseStructLiteralPtr) Log(ctx context.Context, level logrus.Level, msg string) {
	tc.logger.WithContext(ctx).Log(level, msg)
}