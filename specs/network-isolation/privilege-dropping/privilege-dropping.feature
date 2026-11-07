@domain:network-isolation @area:privilege-dropping @from:offline/capabilities
Feature: Stripping privileges from the sandboxed program
  Operators who run untrusted programs under offline rely on the program holding no capabilities, so it cannot reconfigure or escape the network isolation around it.

  Background:
    Given the operator runs a program under offline

  @id:network-isolation.privilege-dropping.001 @rule:network-isolation.privilege-dropping.no-capabilities-in-sandbox @verified @source:capabilities.go#dropCapabilities @test:offline_test.go#TestCapabilitySweepCoversKernelRange @questionable
  Scenario: The sandboxed program holds no capability it could use to undo the isolation
    Given the running kernel defines capabilities up to its highest capability number
    When the operator runs "offline <program>"
    Then the program's capability bounding set is empty, for every capability the running kernel defines
    And the program's ambient capability set is empty
    And running a setuid or file-capability binary inside the sandbox gives the program no capability back

  @id:network-isolation.privilege-dropping.002 @rule:network-isolation.privilege-dropping.drop-failure-not-reported @inferred @source:capabilities.go#dropCapabilities @questionable
  Scenario: A capability that fails to drop does not stop the program and is not reported
    Given the kernel refuses to drop one of the program's capabilities
    When the operator runs "offline <program>"
    Then the program still starts inside the sandbox
    And offline prints no warning or error about the capability on stderr
    And offline's exit code is the program's own exit code
