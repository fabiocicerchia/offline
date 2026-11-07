@domain:sandbox-options @area:blocked-call-logging @from:offline/seccomp
Feature: Reporting blocked network calls
  Serves operators diagnosing why a sandboxed program fails: with one flag, each refused network attempt is named on stderr, and the attempt is still refused.

  Background:
    Given Given the operator runs a program under offline

  @id:sandbox-options.blocked-call-logging.001 @rule:sandbox.logging.blocked-calls-reported-and-denied @verified @source:seccomp.go#logBlockedSyscalls @test:offline_test.go#TestKernelEnforcesFilter @test:offline_test.go#TestFilterDeniesTheNetwork
  Scenario: With logging on, each refused network call is reported and still refused
    Given the operator passed --log-external
    When the wrapped program, running as pid 4242, opens an IPv4 socket
    Then stderr shows "offline: blocked socket (pid 4242)"
    And the wrapped program gets "operation not permitted"
