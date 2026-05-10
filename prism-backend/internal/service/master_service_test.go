package service

import (
	"testing"

	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestValidateLenderCountryRules(t *testing.T) {
	countryID := "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name    string
		req     model.CreateLenderRequest
		wantErr bool
	}{
		{
			name: "bilateral requires country",
			req: model.CreateLenderRequest{
				Name: "Japan International Cooperation Agency",
				Type: "Bilateral",
			},
			wantErr: true,
		},
		{
			name: "ksa allows empty country",
			req: model.CreateLenderRequest{
				Name: "Kuwait Fund",
				Type: "KSA",
			},
		},
		{
			name: "ksa allows country",
			req: model.CreateLenderRequest{
				CountryID: &countryID,
				Name:      "Saudi Fund for Development",
				Type:      "KSA",
			},
		},
		{
			name: "multilateral rejects country",
			req: model.CreateLenderRequest{
				CountryID: &countryID,
				Name:      "World Bank",
				Type:      "Multilateral",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLender(tt.req)
			if tt.wantErr && err == nil {
				t.Fatal("validateLender() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateLender() error = %v, want nil", err)
			}
		})
	}
}
