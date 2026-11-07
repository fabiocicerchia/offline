@domain:installation @area:local-install @from:offline/makefile
Feature: Installing and uninstalling offline locally
  Lets an operator put the offline binary and its man page on their machine, per user by default or under a system prefix, and take them away again.

  Background:
    Given the operator has a checkout of offline with the Go toolchain and the seccomp library available

  @id:installation.local-install.001 @rule:installation.local-install.per-user-default @inferred @source:Makefile#install
  Scenario Outline: Without a prefix, offline installs into the operator's own ~/.local
    Given no install prefix is chosen
    When the operator runs make install <overrides>
    Then the offline binary is built from the checkout
    And the binary is placed at <binary path>
    And the man page is placed at ~/.local/share/man/man1/offline.1
    And the operator sees "installed <binary path> and ~/.local/share/man/man1/offline.1"

    Examples:
      | overrides             | binary path            |
      | (none)                | ~/.local/bin/offline   |
      | BINDIR=/opt/tools/bin | /opt/tools/bin/offline |

  @id:installation.local-install.002 @rule:installation.local-install.system-prefix @inferred @source:Makefile#install
  Scenario Outline: With a prefix, offline installs under that prefix, optionally staged below a destination root
    Given the operator chooses the install prefix /usr/local
    When the operator runs make install PREFIX=/usr/local <staging>
    Then the binary is placed at <root>/usr/local/bin/offline, executable by everyone
    And the man page is placed at <root>/usr/local/share/man/man1/offline.1, readable by everyone
    And the operator sees "installed <root>/usr/local/bin/offline"

    Examples:
      | staging              | root         |
      | (none)               |              |
      | DESTDIR=/tmp/pkgroot | /tmp/pkgroot |

  @id:installation.local-install.003 @rule:installation.local-install.uninstall-removes-installed-files @inferred @source:Makefile#uninstall
  Scenario Outline: Uninstalling removes the binary and man page that install put down
    Given offline was installed with <install options>
    When the operator runs make uninstall <install options>
    Then <binary path> no longer exists
    And <man path> no longer exists
    And uninstalling again when nothing is installed still succeeds

    Examples:
      | install options                  | binary path                  | man path                                  |
      | (none)                           | ~/.local/bin/offline         | ~/.local/share/man/man1/offline.1         |
      | PREFIX=/usr/local                | /usr/local/bin/offline       | /usr/local/share/man/man1/offline.1       |
      | PREFIX=/usr DESTDIR=/tmp/pkgroot | /tmp/pkgroot/usr/bin/offline | /tmp/pkgroot/usr/share/man/man1/offline.1 |
