@domain:sandbox-options @area:policy-handoff @from:offline/offline
Feature: Carrying the chosen options into the sandbox
  Ensures that the options an operator types are exactly the options the isolated stage enforces, no weaker and no stronger.

  Background:
    Given an operator running a program under offline

  @id:sandbox-options.policy-handoff.001 @rule:sandbox-options.policy-handoff.options-reach-isolated-stage @verified @source:offline.go#sandboxFromEnv @test:offline_test.go#TestSandboxSurvivesTheEnvRoundTrip
  Scenario Outline: Every combination of options is enforced inside the sandbox
    Given the operator chose the options <options>
    When the operator runs a program under offline
    Then loopback is <loopback> inside the sandbox
    And refused network calls are <logging> on stderr

    Examples:
      | options                        | loopback | logging      |
      | none                           | down     | not reported |
      | --keep-loopback                | up       | not reported |
      | --log-external                 | down     | reported     |
      | --keep-loopback --log-external | up       | reported     |

  @id:sandbox-options.policy-handoff.002 @rule:sandbox-options.policy-handoff.inherited-markers-relax-sandbox @inferred @source:offline.go#reexecIsolated @questionable
  Scenario: Option markers already in the operator's environment silently relax the sandbox
    Given the operator's shell environment already has _AIRGAP_KEEP_LOOPBACK set to 1
    When the operator runs "curl http://127.0.0.1:8080" under offline without --keep-loopback
    Then IP sockets are allowed inside the sandbox
    And loopback is brought up as if --keep-loopback had been given
    And nothing on stderr says that the sandbox was relaxed
