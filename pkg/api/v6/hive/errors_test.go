package hive

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	type fields struct {
		Code    string
		Message string
		Op      string
		Err     error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Err", fields{Err: errors.New("example")}, "example"},
		{"OpMessage", fields{Op: "testing", Message: "error"}, "testing: error"},
		{"CodeMessage", fields{Code: ErrInternal, Message: "error"}, "<INTERNAL> error"},
		{"OpCodeMessage", fields{Op: "testing", Code: ErrInternal, Message: "error"},
			"testing: <INTERNAL> error"},
		{"OpCodeMessageError", fields{Op: "testing", Code: ErrInternal, Message: "error", Err: errors.New("example")},
			"testing: example"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Error{
				Code:    tt.fields.Code,
				Message: tt.fields.Message,
				Op:      tt.fields.Op,
				Err:     tt.fields.Err,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorCode(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Nil", args{err: nil}, ""},
		{"Standard", args{err: errors.New("example")}, ErrInternal},
		{"Error", args{err: &Error{Code: ErrMissingParameter}}, ErrMissingParameter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorCode(tt.args.err); got != tt.want {
				t.Errorf("ErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Nil", args{err: nil}, ""},
		{"Standard", args{err: errors.New("example")}, "internal error has occurred"},
		{"Error", args{err: &Error{Message: "message"}}, "message"},
		{"NestedError", args{err: &Error{Err: &Error{Message: "message"}}}, "message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorMessage(tt.args.err); got != tt.want {
				t.Errorf("ErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
