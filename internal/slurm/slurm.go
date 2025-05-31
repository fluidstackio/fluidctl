package slurm

import (
	"fmt"
	"net/http"

	client "github.com/fluidstackio/atlas-client-go/v1alpha1"
	"github.com/fluidstackio/fluidctl/internal/auth"
	"github.com/fluidstackio/fluidctl/internal/format"
	"github.com/fluidstackio/fluidctl/internal/utils"
	"github.com/google/uuid"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "slurm",
		Short: "Manage slurm",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.PersistentFlags().StringP("project", "P", "", "Project ID")

	cmd.AddCommand(
		ClusterCommand(),
	)

	return &cmd
}

func ClusterCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "clusters",
		Short: "Manage slurm clusters",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		ListCommand(),
	)

	return &cmd
}

func ListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "list slurm clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")
			projectID, err := uuid.Parse(utils.MustGetStringFlag(cmd, "project"))
			if err != nil {
				return fmt.Errorf("invalid project ID: %w", err)
			}

			token, err := auth.Login(cmd)
			if err != nil {
				return fmt.Errorf("failed to login: %w", err)
			}
			bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(token)
			if err != nil {
				return fmt.Errorf("failed to create bearer auth: %w", err)
			}

			c, err := client.NewClientWithResponses(url+"/api/v1alpha1/", client.WithRequestEditorFn(bearerAuth.Intercept))
			if err != nil {
				return err
			}

			res, err := c.GetSlurmClustersWithResponse(cmd.Context(), &client.GetSlurmClustersParams{
				XPROJECTID: projectID,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to list instances: %s", res.Status())
			}

			f := utils.MustGetStringFlag(cmd, "format")
			m, err := format.NewMarshaller(format.Format(f))
			if err != nil {
				return err
			}

			b, err := m.Marshal(res.JSON200)
			if err != nil {
				return err
			}

			fmt.Println(string(b))

			return nil
		},
	}

	return &cmd
}
