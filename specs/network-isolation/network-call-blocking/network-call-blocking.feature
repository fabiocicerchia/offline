@domain:network-isolation @area:network-call-blocking @from:offline/seccomp
Feature: Failing network calls immediately
  Serves operators who need a network attempt to fail at once with a readable error rather than hang until a timeout, and who need the run to stop if that protection cannot be put in place.

  Background:
    Given Given the operator runs a program under offline

  @id:network-isolation.network-call-blocking.001 @rule:isolation.calls.network-calls-refused-by-default @verified @source:seccomp.go#blockNetworkCalls @test:offline_test.go#TestFilterDeniesTheNetwork @test:offline_test.go#TestSandboxedProgram
  Scenario Outline: Network calls fail at once with a permission error by default
    Given the operator passed no sandbox flags
    When the wrapped program makes a <call> call
    Then the call fails immediately instead of waiting for a timeout
    And the wrapped program gets "operation not permitted"

    Examples:
      | call     |
      | connect  |
      | bind     |
      | listen   |
      | accept   |
      | accept4  |
      | sendto   |
      | sendmsg  |
      | recvfrom |
      | recvmsg  |

  @id:network-isolation.network-call-blocking.002 @rule:isolation.calls.local-socket-connections-refused @inferred @source:seccomp.go#blockNetworkCalls @questionable
  Scenario: Local Unix-socket connections are refused too by default
    Given the operator passed no sandbox flags
    And a local service is listening on a Unix socket the program can see
    When the wrapped program connects to that Unix socket
    Then the connection fails with "operation not permitted"

  @id:network-isolation.network-call-blocking.003 @rule:isolation.calls.no-run-without-filter @inferred @source:seccomp.go#installSeccomp
  Scenario: The program does not run if the network filter cannot be put in place
    Given the network filter cannot be loaded on this host
    When the operator runs a program under offline
    Then the wrapped program is never started
    And offline exits with a failure status and reports the error on stderr
