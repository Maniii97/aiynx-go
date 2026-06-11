package models_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/Maniii97/aiynx-go/internal/models"
)

func TestDOBValidator(t *testing.T) {
	v := validator.New()
	models.RegisterCustomValidators(v)

	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	today := time.Now().UTC().Format("2006-01-02")
	// 131 years ago: always outside the 130-year rolling window.
	tooOld := time.Now().UTC().AddDate(-131, 0, 0).Format("2006-01-02")

	tests := []struct {
		name    string
		dob     string
		wantErr bool
	}{
		{name: "valid past date", dob: "1990-05-10", wantErr: false},
		{name: "today is valid", dob: today, wantErr: false},
		{name: "future date rejected", dob: tomorrow, wantErr: true},
		{name: "over 130 years old rejected", dob: tooOld, wantErr: true},
		{name: "bad format dd-mm-yyyy", dob: "10-05-1990", wantErr: true},
		{name: "bad format dd/mm/yyyy", dob: "10/05/1990", wantErr: true},
		{name: "non-date string", dob: "not-a-date", wantErr: true},
		{name: "empty string", dob: "", wantErr: true},
		{name: "invalid day", dob: "1990-02-30", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := models.CreateUserRequest{Name: "Test User", DOB: tc.dob}
			err := v.Struct(req)
			if (err != nil) != tc.wantErr {
				t.Errorf("DOB=%q: got error=%v, wantErr=%v", tc.dob, err, tc.wantErr)
			}
		})
	}
}
