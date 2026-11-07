@domain:program-execution @area:command-line @from:offline/offline
Feature: Invoking offline from the command line
  Gives the operator a predictable command line: options first, then the program and its arguments untouched, with help and refusals they can script against.

  Background:
    Given an operator at a shell with offline installed

  @id:program-execution.command-line.001 @rule:program-execution.command-line.usage-and-refusals @verified @source:offline.go#main @test:offline_test.go#TestCommandLine
  Scenario Outline: Help, a missing program and an unknown option each get their own answer
    Given the operator types offline <arguments>
    When the operator runs the command
    Then stderr shows <message>
    And offline exits with status <status>

    Examples:
      | arguments           | message                                                             | status |
      | (nothing)           | the usage line and both options, --keep-loopback and --log-external | 1      |
      | --help              | the usage line and both options, --keep-loopback and --log-external | 0      |
      | --no-such-flag true | "flag provided but not defined" followed by the usage               | 2      |

  @id:program-execution.command-line.002 @rule:program-execution.command-line.args-after-program-untouched @verified @source:offline.go#reexecIsolated @test:offline_test.go#TestSandboxedProgram
  Scenario: Everything after the program name goes to the program, not to offline
    Given the operator wants to run "sh -c 'exit 7'"
    When the operator runs offline sh -c 'exit 7'
    Then sh receives -c and 'exit 7' as its own arguments
    And offline exits with status 7
