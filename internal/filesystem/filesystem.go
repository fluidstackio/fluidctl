package filesystem

import (
	"fmt"
	"net/http"

	atlas "github.com/fluidstackio/atlas-client-go/v1alpha1"
	"github.com/fluidstackio/fluidctl/internal/auth"
	"github.com/fluidstackio/fluidctl/internal/format"
	"github.com/fluidstackio/fluidctl/internal/utils"
	"github.com/google/uuid"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filesystems",
		Short: "Manage filesystems",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.PersistentFlags().StringP("project", "P", "", "Project ID")

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
		Short: "Create a new filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")
			name := utils.MustGetStringFlag(cmd, "name")
			size := utils.MustGetStringFlag(cmd, "size")

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

			c, err := atlas.NewClientWithResponses(url+"/api/v1alpha1/", atlas.WithRequestEditorFn(bearerAuth.Intercept))
			if err != nil {
				return err
			}

			res, err := c.PostFilesystemsWithResponse(cmd.Context(), &atlas.PostFilesystemsParams{
				XPROJECTID: projectID,
			}, atlas.FilesystemsPostRequest{
				Name: name,
				Size: size,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusCreated {
				return fmt.Errorf("failed to create filesystem: %s", res.Status())
			}

			return nil
		},
	}

	cmd.Flags().String("name", "", "Name of the filesystem")
	cmd.Flags().String("size", "1024Gi", "Size of the filesystem in GiB")

	return cmd
}

func DeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")
			id, err := uuid.Parse(utils.MustGetStringFlag(cmd, "id"))
			if err != nil {
				return fmt.Errorf("invalid UUID: %w", err)
			}

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

			c, err := atlas.NewClientWithResponses(url+"/api/v1alpha1/", atlas.WithRequestEditorFn(bearerAuth.Intercept))
			if err != nil {
				return err
			}

			res, err := c.DeleteFilesystemsIdWithResponse(cmd.Context(), id, &atlas.DeleteFilesystemsIdParams{
				XPROJECTID: projectID,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusNoContent {
				return fmt.Errorf("failed to delete filesystem: %s", res.Status())
			}

			fmt.Printf("Deleting filesystem with ID: %s\n", id)

			return nil
		},
	}

	cmd.Flags().String("id", "", "Filesystem ID")

	return cmd
}

func ListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all filesystems",
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

			c, err := atlas.NewClientWithResponses(url+"/api/v1alpha1/", atlas.WithRequestEditorFn(bearerAuth.Intercept))
			if err != nil {
				return err
			}

			res, err := c.GetFilesystemsWithResponse(cmd.Context(), &atlas.GetFilesystemsParams{
				XPROJECTID: projectID,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to list filesystems: %s", res.Status())
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
		Short: "Get details of a filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")
			projectID, err := uuid.Parse(utils.MustGetStringFlag(cmd, "project"))
			if err != nil {
				return fmt.Errorf("invalid project ID: %w", err)
			}

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

			c, err := atlas.NewClientWithResponses(url+"/api/v1alpha1/", atlas.WithRequestEditorFn(bearerAuth.Intercept))
			if err != nil {
				return err
			}

			res, err := c.GetFilesystemsIdWithResponse(cmd.Context(), id, &atlas.GetFilesystemsIdParams{
				XPROJECTID: projectID,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to get filesystem: %s", res.Status())
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

	cmd.Flags().String("id", "", "Filesystem ID")

	return cmd
}
