@domain:installation @area:container-image @from:offline/dockerfile
Feature: Running offline from the container image
  Lets an operator run offline from a published image without installing Go or libseccomp on the host, with the image itself not running as root.

  Background:
    Given the operator has built the offline container image

  @id:installation.container-image.001 @rule:installation.container-image.entrypoint-is-offline @inferred @source:Dockerfile#ENTRYPOINT
  Scenario: Arguments given to the container are passed to offline
    Given the container is started with the namespace permissions offline needs
    When the operator runs the image with "curl https://example.com" as its arguments
    Then offline itself is the command that runs inside the container
    And offline receives "curl https://example.com" as the program and its arguments to sandbox

  @id:installation.container-image.002 @rule:installation.container-image.runs-as-non-root @inferred @source:Dockerfile#USER
  Scenario: The container image runs offline as an unprivileged user, not root
    When the operator runs the image without overriding its user
    Then offline runs as user ID 10001
    And offline does not run as root (user ID 0) on the host side of the container
