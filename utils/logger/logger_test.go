/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2025 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package logger

import (
	"context"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/consts"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetContextLoggerContext(t *testing.T) {
	requestID := "myRequestID"
	ctx := context.WithValue(context.Background(), consts.RequestIDLabel, requestID) // nolint:staticcheck
	ctxLogger, err := GetZapContextLogger(ctx)
	if err != nil {
		t.Errorf("Got error from GetLoggerWithContext: %s", err)
	}

	ctxLogger.Info("TestGetZapContextLoggerContext")
}

func TestGetContextLoggerNullContext(t *testing.T) {
	ctxLogger, err := GetZapContextLogger(context.Background())
	if err != nil {
		t.Errorf("Got error from GetLoggerWithContext: %s", err)
	}

	ctxLogger.Info("TestGetZapContextLoggerNullContext")
}

func TestGetDefaultContextLogger(t *testing.T) {
	ctxLogger, err := GetZapDefaultContextLogger()
	if err != nil {
		t.Errorf("Got error from GetLoggerWithContext: %s", err)
	}

	ctxLogger.Info("TestGetDefaultContextLogger")
}

func TestGetContextLoggerFromLoggerContext(t *testing.T) {
	requestID := "myRequestID"
	ctx := context.WithValue(context.Background(), consts.RequestIDLabel, requestID) // nolint:staticcheck
	parentLogger, _ := GetZapLogger()
	ctxLogger, err := GetZapContextLoggerFromLogger(ctx, parentLogger)
	if err != nil {
		t.Errorf("Got error from GetLoggerWithContext: %s", err)
	}

	ctxLogger.Info("TestGetZapContextLoggerContext")
}

func TestGetContextLoggerFromLoggerNullContext(t *testing.T) {
	parentLogger, _ := GetZapLogger()
	ctxLogger, err := GetZapContextLoggerFromLogger(context.Background(), parentLogger)
	if err != nil {
		t.Errorf("Got error from GetLoggerWithContext: %s", err)
	}

	ctxLogger.Info("TestGetZapContextLoggerNullContext")
}

func TestGetContextLoggerFromLoggerNullLogger(t *testing.T) {
	requestID := "myRequestID"
	ctx := context.WithValue(context.Background(), consts.RequestIDLabel, requestID) // nolint:staticcheck
	_, err := GetZapContextLoggerFromLogger(ctx, nil)
	assert.Equal(t, "a valid logger needs to be passed in", err.Error())
}

func TestGetContextLoggerFromLoggerNullLoggerAndContext(t *testing.T) {
	requestID := "myRequestID"
	ctx := context.WithValue(context.Background(), consts.RequestIDLabel, requestID) // nolint:staticcheck
	_, err := GetZapContextLoggerFromLogger(ctx, nil)
	assert.Equal(t, "a valid logger needs to be passed in", err.Error())
}

func TestAddContextFields(t *testing.T) {
	requestID := "myRequestID"
	ctx := context.WithValue(context.Background(), consts.RequestIDLabel, requestID) // nolint:staticcheck
	triggerKey := "myTriggerKey"
	ctx = context.WithValue(ctx, consts.TriggerKeyLabel, triggerKey) // nolint:staticcheck
	parentLogger, _ := GetZapLogger()
	ctxLogger := addContextFields(ctx, parentLogger)
	ctxLogger.Info("TestAddContextFields")
}

// Tests to make sure only the request ID is added
func TestAddContextFieldsTestValue(t *testing.T) {
	requestID := "myRequestID"
	type key string
	var testLabel key = "testLabel"
	ctx := context.WithValue(context.Background(), consts.RequestIDLabel, requestID) // nolint:staticcheck
	ctx = context.WithValue(ctx, testLabel, "test")
	parentLogger, _ := GetZapLogger()
	ctxLogger := addContextFields(ctx, parentLogger)
	ctxLogger.Info("TestAddContextFields")
}

func TestCreateRequestIdField(t *testing.T) {
	requestID := "myRequestID"
	ctx := context.WithValue(context.Background(), consts.RequestIDLabel, requestID) // nolint:staticcheck
	field := CreateZapRequestIDField(ctx)
	if field.Key != consts.RequestIDLabel {
		t.Errorf("Expected key value to be: %s", consts.RequestIDLabel)
	}
	if field.String != requestID {
		t.Errorf("Expected value to be %s", requestID)
	}
	globalLogger, err := GetZapLogger()
	if err != nil {
		t.Errorf("GetZapLogger failed to create a logger")
	}
	globalLogger.Info("TestCreateRequestIdField", field)
}

func TestCreateRequestIdFieldNullContext(t *testing.T) {
	field := CreateZapRequestIDField(context.Background())
	if field.Key != consts.RequestIDLabel {
		t.Errorf("Expected key value to be: %s", consts.RequestIDLabel)
	}
	if field.String != "" {
		t.Errorf("Expected value to be %s", "")
	}
	globalLogger, err := GetZapLogger()
	if err != nil {
		t.Errorf("GetZapLogger failed to create a logger")
	}
	globalLogger.Info("TestCreateZapRequestIDFieldNullContext", field)
}

func TestCreateZapRequestIDFieldNoRequestID(t *testing.T) {
	field := CreateZapRequestIDField(context.Background())
	if field.Key != consts.RequestIDLabel {
		t.Errorf("Expected key value to be: %s", consts.RequestIDLabel)
	}
	if field.String != "" {
		t.Errorf("Expected value to be %s", "")
	}
	globalLogger, err := GetZapLogger()
	if err != nil {
		t.Errorf("GetZapLogger failed to create a logger")
	}
	globalLogger.Info("TestCreateZapRequestIDFieldNoRequestID", field)
}

func TestCreateTriggerKeyField(t *testing.T) {
	triggerKey := "myTriggerKey"
	ctx := context.WithValue(context.Background(), consts.TriggerKeyLabel, triggerKey) // nolint:staticcheck
	field := CreateZapTiggerKeyField(ctx)
	if field.Key != consts.TriggerKeyLabel {
		t.Errorf("Expected key value to be: %s", consts.TriggerKeyLabel)
	}
	if field.String != triggerKey {
		t.Errorf("Expected value to be %s", triggerKey)
	}
	globalLogger, err := GetZapLogger()
	if err != nil {
		t.Errorf("GetZapLogger failed to create a logger")
	}
	globalLogger.Info("TestCreateTriggerKeyField", field)
}

func TestCreateTriggerKeyFieldNullContext(t *testing.T) {
	field := CreateZapTiggerKeyField(context.Background())
	if field.Key != consts.TriggerKeyLabel {
		t.Errorf("Expected key value to be: %s", consts.TriggerKeyLabel)
	}
	if field.String != "" {
		t.Errorf("Expected value to be %s", "")
	}
	globalLogger, err := GetZapLogger()
	if err != nil {
		t.Errorf("GetZapLogger failed to create a logger")
	}
	globalLogger.Info("TestCreateTriggerKeyFieldNullContext", field)
}

func TestCreateZapNoTriggerKeyField(t *testing.T) {
	field := CreateZapTiggerKeyField(context.Background())
	if field.Key != consts.TriggerKeyLabel {
		t.Errorf("Expected key value to be: %s", consts.TriggerKeyLabel)
	}
	if field.String != "" {
		t.Errorf("Expected value to be %s", "")
	}
	globalLogger, err := GetZapLogger()
	if err != nil {
		t.Errorf("GetZapLogger failed to create a logger")
	}
	globalLogger.Info("TestCreateZapNoTriggerKeyField", field)
}

func TestCreatePodNameLoggerNotSet(t *testing.T) {
	logger, err := GetZapLogger()
	if err != nil {
		t.Errorf("Got error when creating global logger: %s", err.Error())
	}
	logger, err = CreatePodNameLogger(logger)
	if err != nil {
		t.Errorf("Got error when creating pod name logger: %s", err.Error())
	}
	logger.Info("TestCreatePodNameLoggerNotSet")
}

func TestCreatePodNameLogger(t *testing.T) {
	_ = os.Setenv(consts.PodNameEnvVar, "myPodName")
	logger, err := GetZapLogger()
	if err != nil {
		t.Errorf("Got error when creating global logger: %s", err.Error())
	}
	logger, err = CreatePodNameLogger(logger)
	if err != nil {
		t.Errorf("Got error when creating pod name logger: %s", err.Error())
	}
	logger.Info("TestCreatePodNameLogger")
}

func TestCreatePodNameLoggerNilLogger(t *testing.T) {
	_, err := CreatePodNameLogger(nil)
	if err == nil {
		t.Errorf("Expected to get an error back")
	} else {
		assert.Equal(t, "logger passed in can not be null", err.Error())
	}
}

func TestCreatePodNameKeyField(t *testing.T) {
	podNameValue := "myPodName"
	_ = os.Setenv(consts.PodNameEnvVar, podNameValue)
	field := CreateZapPodNameKeyField()
	if field.Key != PodName {
		t.Errorf("Expected key value to be: %s", PodName)
	}
	if field.String != podNameValue {
		t.Errorf("Expected value to be %s", podNameValue)
	}
}

func TestCreatePodNameKeyFieldNotSet(t *testing.T) {
	_ = os.Unsetenv(consts.PodNameEnvVar)
	field := CreateZapPodNameKeyField()
	if field.Key != PodName {
		t.Errorf("Expected key value to be: %s", PodName)
	}
	if field.String != "" {
		t.Errorf("Expected value to be %s", "")
	}
}

func TestGenerateContextWithRequestID(t *testing.T) {
	ctx := generateContextWithRequestID()
	requestID, ok := ctx.Value(consts.RequestIDLabel).(string)
	assert.True(t, ok)
	assert.NotEqual(t, "", requestID)
}
