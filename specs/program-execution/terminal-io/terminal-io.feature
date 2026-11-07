@domain:program-execution @area:terminal-io @from:offline/run
Feature: Connecting the wrapped program to the terminal
  Operators use offline interactively and in pipelines, so the wrapped program must read and write the caller's own terminal streams unchanged.

  Background:
    Given Given offline can create its sandbox on this host

  @id:program-execution.terminal-io.001 @rule:execution.terminal-io.streams-are-the-callers @inferred @source:run.go#wireStdio
  Scenario: The wrapped program reads the caller's input and writes to the caller's output and error streams
    Given Given the operator pipes the text "hello" into offline
    When When the operator runs a program under offline that echoes its input to stdout and writes "warn" to stderr
    Then Then "hello" appears on the operator's stdout
    And And "warn" appears on the operator's stderr
