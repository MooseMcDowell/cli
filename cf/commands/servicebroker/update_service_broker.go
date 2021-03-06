package servicebroker

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type UpdateServiceBroker struct {
	ui     terminal.UI
	config coreconfig.Reader
	repo   api.ServiceBrokerRepository
}

func init() {
	commandregistry.Register(&UpdateServiceBroker{})
}

func (cmd *UpdateServiceBroker) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "update-service-broker",
		Description: T("Update a service broker"),
		Usage: []string{
			T("CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"),
		},
	}
}

func (cmd *UpdateServiceBroker) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_BROKER, USERNAME, PASSWORD, URL as arguments\n\n") + commandregistry.Commands.CommandUsage("update-service-broker"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *UpdateServiceBroker) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.repo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *UpdateServiceBroker) Execute(c flags.FlagContext) error {
	serviceBroker, err := cmd.repo.FindByName(c.Args()[0])
	if err != nil {
		return err
	}

	cmd.ui.Say(T("Updating service broker {{.Name}} as {{.Username}}...",
		map[string]interface{}{
			"Name":     terminal.EntityNameColor(serviceBroker.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	serviceBroker.Username = c.Args()[1]
	serviceBroker.Password = c.Args()[2]
	serviceBroker.URL = c.Args()[3]

	err = cmd.repo.Update(serviceBroker)

	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
