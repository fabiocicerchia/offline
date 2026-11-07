@domain:developer-tooling @area:contributor-setup @from:offline/makefile
Feature: Discovering the available make targets
  Shows a contributor every make verb the repo offers, grouped, so they need no knowledge outside the Makefile.

  @id:developer-tooling.contributor-setup.001 @rule:developer-tooling.help.lists-targets-by-default @inferred @source:Makefile#help @questionable
  Scenario: Running make with no target lists the documented targets, but leaves out install and uninstall
    Given a contributor is in the offline checkout
    When the contributor runs make with no target
    Then the groups General, Build and Quality are shown as headings
    And help, setup, build, lint, test, run, format and analyze are listed each with its one-line description
    And install and uninstall are not listed
    And nothing is built or run
