@domain:network-isolation @area:namespaces @from:offline/offline
Feature: Running a program in its own isolated namespaces
  Lets an operator run any program cut off from the host network and host IPC without needing root, which is the core promise of offline.

  Background:
    Given an operator on a Linux host that allows unprivileged user namespaces

  @id:network-isolation.namespaces.001 @rule:network-isolation.namespaces.private-namespace-set @verified @source:offline.go#jailAttr @test:offline_test.go#TestSandboxedProgram
  Scenario: The wrapped program gets its own network, mount, IPC and UTS namespaces
    Given a program that lists the network interfaces it can see
    When the operator runs the program under offline
    Then the program sees no host network interface
    And the program cannot reach the host's IPC objects
    And mounts made inside the sandbox do not appear on the host
    And changing the hostname inside the sandbox leaves the host's hostname unchanged

  @id:network-isolation.namespaces.002 @rule:network-isolation.namespaces.caller-is-root-inside @inferred @source:offline.go#singleIDMap
  Scenario: The caller is root inside the sandbox and nobody on the host
    Given the operator's own user ID is 1000
    When the operator runs "id -u" under offline
    Then the program prints 0
    And the program cannot change its supplementary groups

  @id:network-isolation.namespaces.003 @rule:network-isolation.namespaces.host-process-view-kept @inferred @source:offline.go#jailNamespaces
  Scenario: The wrapped program keeps the host's process numbering
    Given a runtime that reads its own entry under /proc at startup, such as Node
    When the operator runs that runtime under offline
    Then the runtime's own process ID matches its entry under /proc
    And the runtime starts normally instead of reading an unrelated host process

  @id:network-isolation.namespaces.004 @rule:network-isolation.namespaces.fails-closed-without-namespaces @inferred @source:offline.go#reexecIsolated
  Scenario: A host that refuses the sandbox never runs the program unconfined
    Given a host that refuses unprivileged user namespaces
    When the operator runs "curl https://example.com" under offline
    Then curl is never started
    And offline prints the system error on stderr
    And offline exits with status 1
