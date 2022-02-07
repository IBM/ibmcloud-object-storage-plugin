/******************************************************************************r
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package logger

import (
	"context"
	"errors"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/consts"
	uid "github.com/gofrs/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	// PodName is the zap field key label for pod name
	PodName = "podName"
)

// ZapLogger is the global logger
var ZapLogger *zap.Logger

// GetZapLogger returns an instance of the logger, initializing a new logger
func GetZapLogger() (*zap.Logger, error) {
	if ZapLogger == nil {
		return NewZapLogger()
	}
	return ZapLogger, nil
}

// GetZapContextLogger Creates a new logger based from the global logger and adds values from the
// context as logging fields. If the context passed in is null then it
// returns the global logger
func GetZapContextLogger(ctx context.Context) (*zap.Logger, error) {
	var contextLogger *zap.Logger
	globalLogger, _ := GetZapLogger()
	if ctx != nil {
		contextLogger = addContextFields(ctx, globalLogger)
		return contextLogger, nil
	}
	return globalLogger, nil
}

// GetZapDefaultContextLogger Creates a new logger based from the global logger and adds RequestID from the
// context as logging field.
func GetZapDefaultContextLogger() (*zap.Logger, error) {
	var contextLogger *zap.Logger
	globalLogger, _ := GetZapLogger()
	contextLogger = addContextFields(generateContextWithRequestID(), globalLogger)
	return contextLogger, nil
}

// GetZapContextLoggerFromLogger creates a new logger based from an existing logger and adds values from the
// context as logging fields. If the context passed in is null then it
// returns the global logger
func GetZapContextLoggerFromLogger(ctx context.Context, origLogger *zap.Logger) (*zap.Logger, error) {
	var contextLogger *zap.Logger
	if origLogger == nil {
		return origLogger, errors.New("a valid logger needs to be passed in")
	}
	if ctx != nil {
		contextLogger = addContextFields(ctx, origLogger)
		return contextLogger, nil
	}
	return origLogger, nil
}

// Adds fields to an existing logger using values in the context
func addContextFields(ctx context.Context, origLogger *zap.Logger) *zap.Logger {
	if _, ok := ctx.Value(consts.TriggerKeyLabel).(string); ok {
		origLogger = origLogger.With(CreateZapTiggerKeyField(ctx))
	}
	if _, ok := ctx.Value(consts.RequestIDLabel).(string); ok {
		origLogger = origLogger.With(CreateZapRequestIDField(ctx))
	}
	return origLogger
}

// NewZapLogger creates and returns a new global logger. It overwrites the
// existing global logger if that has been previously defined.
func NewZapLogger() (*zap.Logger, error) {
	productionConfig := zap.NewProductionConfig()
	productionConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	ZapLogger, _ := productionConfig.Build()
	return ZapLogger, nil
}

// CreatePodNameLogger takes a zap logger and adds a pod name field
// NOTE: the logger returned is a different logger from the one passed in
func CreatePodNameLogger(logger *zap.Logger) (*zap.Logger, error) {
	if logger == nil {
		return nil, errors.New("logger passed in can not be null")
	}
	podNameField := CreateZapPodNameKeyField()
	return logger.With(podNameField), nil
}

// CreateZapRequestIDField Creates a zap logger field containing the request ID, convenience method for creating the
// field in cases where the ContextLogger can't be used and the field needs to be passed
// in as a parameter in the logging statements
func CreateZapRequestIDField(ctx context.Context) zapcore.Field {
	if ctx != nil {
		if requestID, ok := ctx.Value(consts.RequestIDLabel).(string); ok {
			return zapcore.Field{Key: consts.RequestIDLabel, Type: zapcore.StringType, String: requestID}
		}
	}
	return zapcore.Field{Key: consts.RequestIDLabel, Type: zapcore.StringType, String: ""}
}

// CreateZapTiggerKeyField Creates a zap logger field containing the trigger key for a job, convenience method for creating the
// field in cases where the ContextLogger can't be used and the field needs to be passed
// in as a parameter in the logging statements
func CreateZapTiggerKeyField(ctx context.Context) zapcore.Field {
	if ctx != nil {
		if triggerKey, ok := ctx.Value(consts.TriggerKeyLabel).(string); ok {
			return zapcore.Field{Key: consts.TriggerKeyLabel, Type: zapcore.StringType, String: triggerKey}
		}
	}
	return zapcore.Field{Key: consts.TriggerKeyLabel, Type: zapcore.StringType, String: ""}
}

// CreateZapPodNameKeyField Creates a zap logger field containing the pod name that the container is in,
// convenience method for creating the field so it can be passed
// in as a parameter in the logging statements
func CreateZapPodNameKeyField() zapcore.Field {
	pod := os.Getenv(consts.PodNameEnvVar)
	// if the pod name isn't set then the value will be empty
	return zapcore.Field{Key: PodName, Type: zapcore.StringType, String: pod}
}

// Creates a context that contains a unique request ID
func generateContextWithRequestID() context.Context {
	reqID, _ := uid.NewV4()
	if reqID != nil {
		requestID := reqID.String()
	} else {
		//https://github.com/gofrs/uuid/blob/028e8409cdd0ed11a2b5bb3feb1ae2285ebb94fa/generator.go#L206
		// nil value returned incase of error. so assign a dummy value
		requestID := "mydummyreID"
	}
	return context.WithValue(context.Background(), consts.RequestIDLabel, requestID)
}
