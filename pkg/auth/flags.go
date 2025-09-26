package auth

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AddAuthenticationFlags adds username and password flags to the provided command
func AddAuthenticationFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(UsernameFlagName, "u", "", "Registry username for authentication")
	cmd.Flags().StringP(PasswordFlagName, "p", "", "Registry password for authentication")
}

// AddPersistentAuthenticationFlags adds username and password flags as persistent flags to the provided command
func AddPersistentAuthenticationFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(UsernameFlagName, "u", "", "Registry username for authentication")
	cmd.PersistentFlags().StringP(PasswordFlagName, "p", "", "Registry password for authentication")
}

// ExtractCredentialsFromFlags extracts credentials from command flags
func ExtractCredentialsFromFlags(cmd *cobra.Command) (Credentials, error) {
	username, err := cmd.Flags().GetString(UsernameFlagName)
	if err != nil {
		return Credentials{}, err
	}

	password, err := cmd.Flags().GetString(PasswordFlagName)
	if err != nil {
		return Credentials{}, err
	}

	return NewCredentials(username, password), nil
}

// ExtractCredentialsFromPersistentFlags extracts credentials from persistent command flags
func ExtractCredentialsFromPersistentFlags(cmd *cobra.Command) (Credentials, error) {
	username, err := cmd.PersistentFlags().GetString(UsernameFlagName)
	if err != nil {
		return Credentials{}, err
	}

	password, err := cmd.PersistentFlags().GetString(PasswordFlagName)
	if err != nil {
		return Credentials{}, err
	}

	return NewCredentials(username, password), nil
}

// ValidateFlagCombinations validates that flag combinations make sense
func ValidateFlagCombinations(flags *pflag.FlagSet) error {
	username, _ := flags.GetString(UsernameFlagName)
	password, _ := flags.GetString(PasswordFlagName)

	creds := NewCredentials(username, password)
	validator := &DefaultValidator{}

	return validator.ValidateCredentials(creds)
}

// BindFlagsToConfig binds authentication flags to configuration values
func BindFlagsToConfig(cmd *cobra.Command, config *AuthConfig) error {
	creds, err := ExtractCredentialsFromFlags(cmd)
	if err != nil {
		return err
	}

	config.Credentials = creds
	config.Required = !creds.IsEmpty()

	return nil
}

// BindPersistentFlagsToConfig binds persistent authentication flags to configuration values
func BindPersistentFlagsToConfig(cmd *cobra.Command, config *AuthConfig) error {
	creds, err := ExtractCredentialsFromPersistentFlags(cmd)
	if err != nil {
		return err
	}

	config.Credentials = creds
	config.Required = !creds.IsEmpty()

	return nil
}