@domain:installation @area:one-line-installer @from:offline/install
Feature: Installing offline with the one-line installer
  Lets a user get a working offline binary on their machine with a single curl-piped command, and re-run the same command later to update it.

  Background:
    Given the user's machine has git, make, Go and the seccomp build dependencies

  @id:installation.one-line-installer.001 @rule:installation.installer.fresh-install-clones-and-installs @inferred @source:install.sh#REPO_DIR
  Scenario Outline: A first run fetches offline and installs it
    Given there is no offline checkout at <checkout_dir>
    And OFFLINE_HOME is <offline_home>
    When the user runs the one-line installer
    Then a checkout of github.com/fabiocicerchia/offline exists at <checkout_dir>
    And the offline binary is installed at ~/.local/bin/offline
    And the installer prints "offline installed. Run 'offline <program> [args...]'."

    Examples:
      | offline_home     | checkout_dir           |
      | unset            | ~/.local/share/offline |
      | /opt/src/offline | /opt/src/offline       |

  @id:installation.one-line-installer.002 @rule:installation.installer.updates-existing-checkout @inferred @source:install.sh#REPO_DIR
  Scenario: Re-running the installer updates the existing install in place
    Given an offline checkout already exists at ~/.local/share/offline
    And the upstream main branch has newer commits
    When the user runs the one-line installer
    Then the existing checkout is fast-forwarded to the upstream commits rather than cloned again
    And the rebuilt offline binary replaces the one at ~/.local/bin/offline
    And the installer prints "offline installed. Run 'offline <program> [args...]'."

  @id:installation.one-line-installer.003 @rule:installation.installer.stops-on-failed-step @inferred @source:install.sh#REPO_DIR
  Scenario Outline: The installer stops without installing when it cannot fetch the source
    Given <situation>
    When the user runs the one-line installer
    Then git reports <git_error>
    And the installer exits with a non-zero status
    And no binary is built or installed
    And the success message is not printed

    Examples:
      | situation                                                                                 | git_error                                                 |
      | the checkout at ~/.local/share/offline has local commits that diverged from upstream main | that it cannot fast-forward                               |
      | ~/.local/share/offline is a non-empty directory that is not a git checkout                | that the destination path already exists and is not empty |
