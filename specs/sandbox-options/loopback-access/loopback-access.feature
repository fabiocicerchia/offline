@domain:sandbox-options @area:loopback-access @from:offline/offline
Feature: Keeping loopback reachable
  Lets an operator run programs that need a local server on 127.0.0.1 while still having no route to any other network.

  Background:
    Given an operator running a program under offline

  @id:sandbox-options.loopback-access.001 @rule:sandbox-options.loopback-access.loopback-up-when-kept @verified @source:offline.go#bringUpLoopback @test:offline_test.go#TestSandboxedProgram
  Scenario: With --keep-loopback the program can reach its own local services
    Given a program that listens on 127.0.0.1 and then connects to itself
    When the operator runs that program under offline with --keep-loopback
    Then the connection to 127.0.0.1 succeeds
    And offline exits with status 0

  @id:sandbox-options.loopback-access.002 @rule:sandbox-options.loopback-access.loopback-failure-aborts-before-program @inferred @source:offline.go#runIsolated @questionable
  Scenario: If loopback cannot be brought up the program is not started
    Given a host whose security policy stops the sandbox from configuring its network interface
    When the operator runs a program under offline with --keep-loopback
    Then the program is never started
    And a Go panic message and stack trace appear on stderr
    And offline exits with status 2
