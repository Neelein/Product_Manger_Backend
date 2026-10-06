package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadReadsEnvironmentContract(t *testing.T) {
	c, err := Load(func(key string) string {
		return map[string]string{"PAYMENT_OTP": "test-otp", "DEFAULT_DEPARTMENT_CODE": "7", "PORT": "9000"}[key]
	})
	require.NoError(t, err)
	require.Equal(t, "test-otp", c.PaymentOTP)
	require.Equal(t, 7, c.DefaultDepartmentCode)
	require.Equal(t, "9000", c.Port)
}

func TestLoadRejectsMissingPaymentOTP(t *testing.T) {
	_, err := Load(func(string) string { return "" })
	require.ErrorIs(t, err, ErrMissingPaymentOTP)
}

func TestLoadUsesSafeLocalDefaults(t *testing.T) {
	c, err := Load(func(key string) string {
		if key == "PAYMENT_OTP" {
			return "test-otp"
		}
		return ""
	})
	require.NoError(t, err)
	require.Equal(t, 1, c.DefaultDepartmentCode)
	require.Equal(t, "8090", c.Port)
}
