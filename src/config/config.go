package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var ErrMissingPaymentOTP = errors.New("PAYMENT_OTP is required")

type Config struct {
	DatabaseURL           string
	APIGatewaySecret      string
	PaymentOTP            string
	DefaultDepartmentCode int
	Port                  string
}

func Load(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	code := 1
	if value := strings.TrimSpace(getenv("DEFAULT_DEPARTMENT_CODE")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("DEFAULT_DEPARTMENT_CODE must be a positive integer: %w", err)
		}
		code = parsed
	}
	otp := strings.TrimSpace(getenv("PAYMENT_OTP"))
	if otp == "" {
		return Config{}, ErrMissingPaymentOTP
	}
	port := strings.TrimSpace(getenv("PORT"))
	if port == "" {
		port = "8090"
	}
	return Config{
		DatabaseURL: getenv("DATABASE_URL"), APIGatewaySecret: getenv("API_GATEWAY_SECRET"), PaymentOTP: otp,
		DefaultDepartmentCode: code, Port: port,
	}, nil
}
