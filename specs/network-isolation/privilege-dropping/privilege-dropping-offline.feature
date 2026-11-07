@domain:network-isolation @area:privilege-dropping @from:offline/offline
Feature: Stopping privilege escalation inside the sandbox
  Assures an operator that nothing run inside the sandbox can gain privileges to undo the isolation.

  Background:
    Given an operator running a program under offline

  @id:network-isolation.privilege-dropping.003 @rule:network-isolation.privilege-dropping.no-setuid-escalation @inferred @source:offline.go#runIsolated
  Scenario: Setuid programs gain no privileges inside the sandbox
    Given a setuid-root program such as ping is installed on the host
    When the wrapped program runs that setuid program
    Then the setuid program runs without any extra privileges
    And it cannot lift the sandbox's network restrictions
