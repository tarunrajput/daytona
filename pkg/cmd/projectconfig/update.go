// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var projectConfigUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a project config",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectConfig *apiclient.ProjectConfig
		var res *http.Response
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, false, "Update")
			if projectConfig == nil {
				return nil
			}
		} else {
			projectConfig, res, err = apiClient.ProjectConfigAPI.GetProjectConfig(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if projectConfig == nil {
			return nil
		}

		createDto := []apiclient.CreateProjectDTO{
			{
				Name: projectConfig.Name,
				Source: apiclient.CreateProjectSourceDTO{
					Repository: apiclient.GitRepository{
						Url: projectConfig.RepositoryUrl,
					},
				},
				BuildConfig: projectConfig.BuildConfig,
				EnvVars:     projectConfig.EnvVars,
			},
		}

		projectDefaults := &views_util.ProjectConfigDefaults{
			BuildChoice: views_util.AUTOMATIC,
			Image:       &projectConfig.Image,
			ImageUser:   &projectConfig.User,
		}

		if projectConfig.BuildConfig != nil && projectConfig.BuildConfig.Devcontainer != nil {
			projectDefaults.DevcontainerFilePath = projectConfig.BuildConfig.Devcontainer.FilePath
		}

		create.ProjectsConfigurationChanged, err = create.RunProjectConfiguration(&createDto, *projectDefaults)
		if err != nil {
			return err
		}

		newProjectConfig := apiclient.CreateProjectConfigDTO{
			Name:          projectConfig.Name,
			BuildConfig:   createDto[0].BuildConfig,
			Image:         createDto[0].Image,
			User:          createDto[0].User,
			RepositoryUrl: createDto[0].Source.Repository.Url,
			EnvVars:       createDto[0].EnvVars,
		}

		res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(newProjectConfig).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Project config updated successfully")
		return nil
	},
}
