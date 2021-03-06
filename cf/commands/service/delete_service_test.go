package service_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete-service command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		serviceRepo         *apifakes.FakeServiceRepository
		serviceInstance     models.ServiceInstance
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"yes"},
		}

		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceRepo = new(apifakes.FakeServiceRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-service", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Context("when not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		})

		It("does not pass requirements", func() {
			Expect(runCommand("vestigial-service")).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		It("fails with usage when not provided exactly one arg", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		Context("when the service exists", func() {
			Context("and the service deletion is asynchronous", func() {
				BeforeEach(func() {
					serviceInstance = models.ServiceInstance{}
					serviceInstance.Name = "my-service"
					serviceInstance.GUID = "my-service-guid"
					serviceInstance.LastOperation.Type = "delete"
					serviceInstance.LastOperation.State = "in progress"
					serviceInstance.LastOperation.Description = "delete"
					serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
				})

				Context("when the command is confirmed", func() {
					It("deletes the service", func() {
						runCommand("my-service")

						Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service my-service"}))

						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
							[]string{"OK"},
							[]string{"Delete in progress. Use 'cf services' or 'cf service my-service' to check operation status."},
						))

						Expect(serviceRepo.DeleteServiceArgsForCall(0)).To(Equal(serviceInstance))
					})
				})

				It("skips confirmation when the -f flag is given", func() {
					runCommand("-f", "foo.com")

					Expect(ui.Prompts).To(BeEmpty())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Deleting service", "foo.com"},
						[]string{"OK"},
						[]string{"Delete in progress. Use 'cf services' or 'cf service foo.com' to check operation status."},
					))
				})
			})

			Context("and the service deletion is synchronous", func() {
				BeforeEach(func() {
					serviceInstance = models.ServiceInstance{}
					serviceInstance.Name = "my-service"
					serviceInstance.GUID = "my-service-guid"
					serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
				})

				Context("when the command is confirmed", func() {
					It("deletes the service", func() {
						runCommand("my-service")

						Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service my-service"}))

						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
							[]string{"OK"},
						))

						Expect(serviceRepo.DeleteServiceArgsForCall(0)).To(Equal(serviceInstance))
					})
				})

				It("skips confirmation when the -f flag is given", func() {
					runCommand("-f", "foo.com")

					Expect(ui.Prompts).To(BeEmpty())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Deleting service", "foo.com"},
						[]string{"OK"},
					))
				})
			})
		})

		Context("when the service does not exist", func() {
			BeforeEach(func() {
				serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.NewModelNotFoundError("Service instance", "my-service"))
			})

			It("warns the user the service does not exist", func() {
				runCommand("-f", "my-service")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting service", "my-service"},
					[]string{"OK"},
				))

				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"my-service", "does not exist"}))
			})
		})
	})
})
