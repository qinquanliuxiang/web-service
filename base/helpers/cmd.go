package helpers

import (
	"os"
	"web-service/base/constant"

	"github.com/spf13/cobra"
)

func PreRun(cmd *cobra.Command) {
	if !cmd.Flags().Changed(constant.FlagConfigPath) {
		envConfigPath := os.Getenv(constant.ConfigEnv)
		if envConfigPath != "" {
			cmd.Flags().Set(constant.FlagConfigPath, envConfigPath)
		}
	}

	if !cmd.Flags().Changed(constant.FlagCasbinModePath) {
		envCasbinModePath := os.Getenv(constant.CasbinEnv)
		if envCasbinModePath != "" {
			cmd.Flags().Set(constant.FlagCasbinModePath, envCasbinModePath)
		}
	}
}
