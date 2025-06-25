package scanner

import (
	"testing"
)

func Test_config_Validate_Languages_Config(t *testing.T) {
	err := ValidateLanguagesConfig()
	if err != nil {
		t.Errorf("Test_config_Validate_Languages_Config returned an error: %v", err)
		return
	}
}
