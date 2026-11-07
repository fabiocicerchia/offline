@domain:program-execution @area:exit-status @from:offline/run
Feature: Passing the wrapped program's exit status back
  Scripts and operators that wrap a command in offline rely on getting that command's own exit status back, so offline is a drop-in prefix in pipelines and CI.

  Background:
    Given offline can create its sandbox on this host

  @id:program-execution.exit-status.001 @rule:execution.exit-status.code-passes-through @verified @source:run.go#runAndExit @test:offline_test.go#TestExitCode @test:offline_test.go#TestSandboxedProgram
  Scenario Outline: A program's exit code becomes offline's exit code
    Given a program that exits with code <code>
    When the operator runs the program under offline
    Then offline exits with code <code>
    And offline adds nothing of its own to stderr

    Examples:
      | code |
      | 0    |
      | 1    |
      | 7    |

  @id:program-execution.exit-status.002 @rule:execution.exit-status.signal-death-is-128-plus-signal @verified @source:run.go#exitCode @test:offline_test.go#TestExitCode
  Scenario Outline: A program killed by a signal makes offline exit with 128 plus the signal number
    Given a program that is killed by <signal>
    When the operator runs the program under offline
    Then offline exits with code <code>

    Examples:
      | signal  | code |
      | SIGTERM | 143  |
      | SIGKILL | 137  |

  @id:program-execution.exit-status.003 @rule:execution.exit-status.unrunnable-program-is-reported @inferred @source:run.go#runAndExit
  Scenario: A program that cannot be started is reported on stderr and offline exits with code 1
    Given no program named "no-such-tool" exists on the PATH
    When the operator runs "no-such-tool" under offline
    Then stderr shows why the program could not be started
    And offline exits with code 1
