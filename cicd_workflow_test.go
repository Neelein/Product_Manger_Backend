package main

import (
	"os"
	"strings"
	"testing"
)

func TestE2EWorkflowProvidesNonProductionPaymentOTP(t *testing.T) {
	workflow, err := os.ReadFile(".github/workflows/cicd.yml")
	if err != nil {
		t.Fatal(err)
	}

	const stepMarker = "      - name: Run E2E tests\n"
	stepStart := strings.Index(string(workflow), stepMarker)
	if stepStart == -1 {
		t.Fatal("E2E workflow must contain a Run E2E tests step")
	}
	nextStep := strings.Index(string(workflow)[stepStart+len(stepMarker):], "\n      - name:")
	if nextStep == -1 {
		nextStep = len(string(workflow)) - stepStart - len(stepMarker)
	}
	e2eStep := string(workflow)[stepStart : stepStart+len(stepMarker)+nextStep]
	const expected = "          PAYMENT_OTP: e2e-payment-otp-placeholder\n"
	if !strings.Contains(e2eStep, expected) {
		t.Fatalf("E2E workflow must set PAYMENT_OTP to the non-production placeholder")
	}
	if strings.Contains(string(workflow), "secrets.PAYMENT_OTP") {
		t.Fatal("E2E workflow must not consume a production payment OTP secret")
	}
}
