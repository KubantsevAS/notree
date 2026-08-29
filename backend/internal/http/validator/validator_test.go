package validator_test

import (
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/validator"
)

type jsonTestPayload struct {
	Data []byte `validate:"json_object"`
}

func TestJSONObject(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Valid JSON object",
			input:   []byte(`{"key": "value", "nested": {"a": 1}}`),
			wantErr: false,
		},
		{
			name:    "Empty byte slice",
			input:   []byte{},
			wantErr: false,
		},
		{
			name:    "Nil byte slice",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "Invalid JSON string",
			input:   []byte(`{not a json}`),
			wantErr: true,
		},
		{
			name:    "Valid JSON but it's an array, not object",
			input:   []byte(`[1, 2, 3]`),
			wantErr: true,
		},
		{
			name:    "Valid JSON but it's a string primitive",
			input:   []byte(`"just a string"`),
			wantErr: true,
		},
		{
			name:    "Valid JSON but it's a number primitive",
			input:   []byte(`123`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := jsonTestPayload{Data: tt.input}
			err := validator.Check(payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type userPayload struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
}

func TestCheckStandardTags(t *testing.T) {
	tests := []struct {
		name    string
		payload userPayload
		wantErr bool
	}{
		{
			name:    "Valid user",
			payload: userPayload{Email: "test@test.com", Age: 25},
			wantErr: false,
		},
		{
			name:    "Missing required email",
			payload: userPayload{Email: "", Age: 25},
			wantErr: true,
		},
		{
			name:    "Invalid email format",
			payload: userPayload{Email: "not-an-email", Age: 25},
			wantErr: true,
		},
		{
			name:    "Age below minimum",
			payload: userPayload{Email: "test@test.com", Age: 17},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Check(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
