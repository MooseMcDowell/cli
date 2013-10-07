package space

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateSpace struct {
	ui        terminal.UI
	spaceRepo api.SpaceRepository
}

func NewCreateSpace(ui terminal.UI, sR api.SpaceRepository) (cmd CreateSpace) {
	cmd.ui = ui
	cmd.spaceRepo = sR
	return
}

func (cmd CreateSpace) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-space")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd CreateSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]
	cmd.ui.Say("Creating space %s...", terminal.EntityNameColor(spaceName))

	apiResponse := cmd.spaceRepo.Create(spaceName)
	if apiResponse.IsNotSuccessful() {
		if apiResponse.ErrorCode == api.SPACE_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Say("Space %s already exists", spaceName)
			return
		}
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\nTIP: Use '%s' to target new space", terminal.CommandColor(cf.Name+" target -s "+spaceName))
}
