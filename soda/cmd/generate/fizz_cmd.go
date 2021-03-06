package generate

import (
	"errors"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/internal/defaults"
	"github.com/spf13/cobra"
)

//FizzCmd generates a new fizz migration
var FizzCmd = &cobra.Command{
	Use:     "fizz [name]",
	Aliases: []string{"migration"},
	Short:   "Generates Up/Down migrations for your database using fizz.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("you must supply a name for your migration")
		}
		cflag := cmd.Flag("path")
		migrationPath := defaults.String(cflag.Value.String(), "./migrations")
		return pop.MigrationCreate(migrationPath, args[0], "fizz", nil, nil)
	},
}
