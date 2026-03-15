package logger

import "context"

type noopLogger struct{}

func NewNoop() Logger { return &noopLogger{} }

func (n *noopLogger) Debug(string, ...interface{})          {}
func (n *noopLogger) Info(string, ...interface{})           {}
func (n *noopLogger) Warn(string, ...interface{})           {}
func (n *noopLogger) Error(string, ...interface{})          {}
func (n *noopLogger) With(...interface{}) Logger            { return n }
func (n *noopLogger) WithContext(context.Context) Logger    { return n }
