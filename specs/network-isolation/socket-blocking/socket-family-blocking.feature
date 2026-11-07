@domain:network-isolation @area:socket-blocking @from:offline/seccomp
Feature: Refusing network sockets
  Serves operators who run untrusted programs offline: the program cannot even create a socket that could reach a network, while local non-network sockets keep working.

  Background:
    Given the operator runs a program under offline

  @id:network-isolation.socket-blocking.001 @rule:isolation.sockets.network-families-refused-by-default @verified @source:seccomp.go#blockSocketFamilies @test:offline_test.go#TestBlockedFamilies @test:offline_test.go#TestFilterDeniesTheNetwork @test:offline_test.go#TestKernelEnforcesFilter @test:offline_test.go#TestSandboxedProgram
  Scenario Outline: IPv4, IPv6 and raw packet sockets are refused by default
    Given the operator passed no sandbox flags
    When the wrapped program opens a <family> socket
    Then the socket is not created
    And the wrapped program gets "operation not permitted"

    Examples:
      | family                 |
      | IPv4 (AF_INET)         |
      | IPv6 (AF_INET6)        |
      | raw packet (AF_PACKET) |

  @id:network-isolation.socket-blocking.002 @rule:isolation.sockets.raw-packet-always-refused @verified @source:seccomp.go#blockSocketFamilies @test:offline_test.go#TestBlockedFamilies @test:offline_test.go#TestFilterDeniesTheNetwork
  Scenario: Raw packet sockets stay refused even when loopback is kept
    Given the operator passed --keep-loopback
    When the wrapped program opens a raw packet (AF_PACKET) socket
    Then the socket is not created
    And the wrapped program gets "operation not permitted"

  @id:network-isolation.socket-blocking.003 @rule:isolation.sockets.local-sockets-creatable @verified @source:seccomp.go#buildSeccompFilter @test:offline_test.go#TestKernelEnforcesFilter
  Scenario: Local Unix sockets can still be created
    Given the operator passed no sandbox flags
    When the wrapped program opens a Unix (AF_UNIX) socket
    Then the socket is created
