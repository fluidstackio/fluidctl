package utils

import (
	"strings"

	"github.com/spf13/cobra"
)

func MustGetStringFlag(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(err)
	}
	return value
}

func MustGetStringSliceFlag(cmd *cobra.Command, name string) []string {
	value, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		panic(err)
	}
	return value
}

func MustGetStringArrayFlag(cmd *cobra.Command, name string) []string {
	value, err := cmd.Flags().GetStringArray(name)
	if err != nil {
		panic(err)
	}
	return value
}

func MustGetBoolFlag(cmd *cobra.Command, name string) bool {
	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(err)
	}
	return value
}

func MustGetIntFlag(cmd *cobra.Command, name string) int {
	value, err := cmd.Flags().GetInt(name)
	if err != nil {
		panic(err)
	}
	return value
}

func ParseAttrs(s string) map[string]string {
	attrs := map[string]string{}
	for _, fs := range strings.Split(s, ",") {
		if k, v, found := strings.Cut(fs, "="); found {
			attrs[k] = v
		} else {
			attrs[fs] = ""
		}
	}

	return attrs
}
