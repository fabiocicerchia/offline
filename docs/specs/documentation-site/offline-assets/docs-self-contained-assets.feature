@domain:documentation-site @area:offline-assets @from:stylesheets/extra
Feature: Rendering the docs without fetching fonts from the network
  Readers and air-gapped operators get docs that render fully from the site itself, with no web-font or CDN request, matching the product's no-network stance.

  @id:docs.documentation-site.offline-assets.001 @rule:docs.assets.system-fonts-only @inferred @source:docs/stylesheets/extra.css#:root
  Scenario: Docs text and code render in the reader's system fonts with no font download
    Given the reader's machine has no access to any font CDN
    When the reader opens a docs page
    Then body text renders in the system sans-serif stack (-apple-system, Segoe UI, Roboto, Helvetica, Arial)
    And code renders in the system monospace stack (ui-monospace, SF Mono, Menlo, Consolas, Liberation Mono)
    And no web font is requested over the network
