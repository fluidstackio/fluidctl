package instance

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	client "github.com/fluidstackio/atlas-client-go/v1alpha1"
	"github.com/fluidstackio/fluidctl/internal/auth"
	"github.com/fluidstackio/fluidctl/internal/format"
	"github.com/fluidstackio/fluidctl/internal/utils"
	"github.com/google/uuid"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func Command() *cobra.Command {
	cmd := cobra.Command{
		Use:   "instances",
		Short: "Manage instances",
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

	return &cmd
}

type UserData struct {
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
}

func CreateCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "create",
		Short: "create instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := utils.MustGetStringFlag(cmd, "url")
			name := utils.MustGetStringFlag(cmd, "name")
			image := utils.MustGetStringFlag(cmd, "image")
			instanceType := utils.MustGetStringFlag(cmd, "type")
			preemptible := utils.MustGetBoolFlag(cmd, "preemptible")

			projectID, err := uuid.Parse(utils.MustGetStringFlag(cmd, "project"))
			if err != nil {
				return fmt.Errorf("invalid project ID: %w", err)
			}

			instance := client.InstancesPostRequest{
				Name:        name,
				Preemptible: &preemptible,
				Type:        instanceType,
			}

			if image != "" {
				instance.Image = &image
			}

			userDataPath := utils.MustGetStringFlag(cmd, "user-data")
			sshAuthrorizedKeyPaths := utils.MustGetStringArrayFlag(cmd, "ssh-authorized-key")

			if userDataPath != "" {
				userData, err := os.ReadFile(userDataPath)
				if err != nil {
					return fmt.Errorf("failed to read user-data file: %w", err)
				}

				if len(sshAuthrorizedKeyPaths) != 0 {
					return fmt.Errorf("cannot specify both user-data and ssh-authorized-key")
				}

				instance.UserData = &userData
			} else {
				sshAuthorizedKeys := []string{}
				for _, sshAuthorizedKeyPath := range sshAuthrorizedKeyPaths {
					sshAuthorizedKey, err := os.ReadFile(sshAuthorizedKeyPath)
					if err != nil {
						return fmt.Errorf("failed to read ssh public-key file: %w", err)
					}

					sshAuthorizedKeys = append(sshAuthorizedKeys, strings.TrimSpace(string(sshAuthorizedKey)))
				}

				b, err := yaml.Marshal(&UserData{
					SSHAuthorizedKeys: sshAuthorizedKeys,
				})
				if err != nil {
					return fmt.Errorf("failed to marshal user-data: %w", err)
				}

				userData := append([]byte("#cloud-config\n"), b...)
				instance.UserData = &userData
			}

			filesystems := []uuid.UUID{}
			for _, fs := range utils.MustGetStringArrayFlag(cmd, "filesystem") {
				id, err := parseFilesystemFlag(fs)
				if err != nil {
					return err
				}

				filesystems = append(filesystems, id)
			}
			if len(filesystems) != 0 {
				instance.Filesystems = &filesystems
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

			res, err := c.PostInstancesWithResponse(cmd.Context(), &client.PostInstancesParams{
				XPROJECTID: projectID,
			}, instance)
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusCreated {
				return fmt.Errorf("failed to create instance: %s", res.Status())
			}

			return nil
		},
	}

	cmd.Flags().String("name", "", "Name of the instance")
	cmd.Flags().String("user-data", "", "Path to cloud-init user-data")
	cmd.Flags().StringArray("ssh-authorized-key", []string{}, "Path to SSH public key")
	cmd.Flags().String("image", "", "Image URL")
	cmd.Flags().StringArray("filesystem", []string{}, "Filesystems to attach (in the format 'id=<UUID>')")
	cmd.Flags().Bool("preemptible", false, "Create a preemptible instance")
	cmd.Flags().String("type", "cpu.2x", "Instance type")

	return &cmd
}

func parseFilesystemFlag(s string) (uuid.UUID, error) {
	attrs := utils.ParseAttrs(s)
	if id, found := attrs["id"]; found {
		res, err := uuid.Parse(id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid filesystem id: %s", id)
		}

		return res, nil
	} else {
		return uuid.Nil, fmt.Errorf("missing 'id' attribute in filesystem: %s", s)
	}
}

func DeleteCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete",
		Short: "delete instance",
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

			c, err := client.NewClientWithResponses(url+"/api/v1alpha1/", client.WithRequestEditorFn(bearerAuth.Intercept))
			if err != nil {
				return err
			}

			res, err := c.DeleteInstancesIdWithResponse(cmd.Context(), id, &client.DeleteInstancesIdParams{
				XPROJECTID: projectID,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusNoContent {
				return fmt.Errorf("failed to delete instance: %s", res.Status())
			}

			fmt.Printf("Deleting instance with ID: %s\n", id)

			return nil
		},
	}

	cmd.Flags().String("id", "", "Instance ID")

	return &cmd
}

func ListCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "list instances",
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

			res, err := c.GetInstancesWithResponse(cmd.Context(), &client.GetInstancesParams{
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

func DescribeCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "describe",
		Short: "describe instance",
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

			c, err := client.NewClientWithResponses(url+"/api/v1alpha1/", client.WithRequestEditorFn(bearerAuth.Intercept))
			if err != nil {
				return err
			}

			res, err := c.GetInstancesIdWithResponse(cmd.Context(), id, &client.GetInstancesIdParams{
				XPROJECTID: projectID,
			})
			if err != nil {
				return err
			}

			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to get instance: %s", res.Status())
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

	cmd.Flags().String("id", "", "Instance ID")

	return &cmd
}
