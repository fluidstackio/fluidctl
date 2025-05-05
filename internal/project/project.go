package project

import (
	"fmt"
	"net/http"

	"github.com/fluidstackio/fluidctl/internal/auth"
	"github.com/fluidstackio/fluidctl/internal/client"
	"github.com/fluidstackio/fluidctl/internal/format"
	"github.com/fluidstackio/fluidctl/internal/utils"
	"github.com/google/uuid"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		CreateCommand(),
		DeleteCommand(),
		ListCommand(),
		DescribeCommand(),
	)

	return cmd
}

func CreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")
			name := utils.MustGetStringFlag(cmd, "name")

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

			res, err := c.PostProjectsWithResponse(cmd.Context(), &client.PostProjectsParams{}, client.ProjectsPostRequest{
				Name: name,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusCreated {
				return fmt.Errorf("failed to create project: %s", res.Status())
			}

			return nil
		},
	}

	cmd.Flags().String("name", "", "Name of the project")

	return cmd
}

func DeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")

			id, err := uuid.Parse(utils.MustGetStringFlag(cmd, "id"))
			if err != nil {
				return fmt.Errorf("invalid UUID: %w", err)
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

			res, err := c.DeleteProjectsIdWithResponse(cmd.Context(), id, &client.DeleteProjectsIdParams{})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusNoContent {
				return fmt.Errorf("failed to delete project: %s", res.Status())
			}

			fmt.Printf("Deleting project with ID: %s\n", id)

			return nil
		},
	}

	cmd.Flags().String("id", "", "Project ID")

	return cmd
}

func ListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")

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

			res, err := c.GetProjectsWithResponse(cmd.Context(), &client.GetProjectsParams{})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to list projects: %s", res.Status())
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
}

func DescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Get details of a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")

			id, err := uuid.Parse(utils.MustGetStringFlag(cmd, "id"))
			if err != nil {
				return fmt.Errorf("invalid UUID: %w", err)
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

			res, err := c.GetProjectsIdWithResponse(cmd.Context(), id, &client.GetProjectsIdParams{})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to get project: %s", res.Status())
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

	cmd.Flags().String("id", "", "Project ID")

	return cmd
}
