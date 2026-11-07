@domain:sandbox-options @area:loopback-access @from:offline/seccomp
Feature: Allowing local loopback traffic
  Serves operators whose program needs to talk to itself over 127.0.0.1/::1: one flag lifts the IP refusals so local traffic works, and the namespace keeps the outside world unreachable.

  Background:
    Given Given the operator runs a program under offline

  @id:sandbox-options.loopback-access.003 @rule:sandbox.loopback.ip-traffic-allowed @verified @source:seccomp.go#buildSeccompFilter @test:offline_test.go#TestFilterDeniesTheNetwork @test:offline_test.go#TestKernelEnforcesFilter @test:offline_test.go#TestSandboxedProgram
  Scenario: With loopback kept, the program can open IP sockets and talk over loopback
    Given the operator passed --keep-loopback
    When the wrapped program listens on 127.0.0.1 and connects to itself
    Then the IPv4 socket is created
    And the connection succeeds
    And no "operation not permitted" error is reported
