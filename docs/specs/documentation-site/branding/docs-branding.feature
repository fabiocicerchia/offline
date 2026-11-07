@domain:documentation-site @area:branding @from:stylesheets/extra
Feature: Presenting the offline brand on the docs and splash page
  Readers landing on the splash page or the rendered docs see one consistent teal identity in light and dark schemes, so the site reads as one product.

  Background:
    Given the offline documentation site is published with its splash page at the site root and the rendered docs under /docs/

  @id:docs.documentation-site.branding.001 @rule:docs.branding.shared-accent-palette @inferred @source:docs/stylesheets/extra.css#:root
  Scenario: The splash page and the docs show the same accent colours
    Given the site's accent pair is green #34d399 and teal #0ea5a5
    When a reader opens the splash page and then the docs
    Then the splash page title is drawn as a gradient from #34d399 to #0ea5a5
    And the docs' primary colour, used for links and highlighted navigation, is teal #0ea5a5
    And changing the accent pair in one place changes both the splash page and the docs

  @id:docs.documentation-site.branding.002 @rule:docs.branding.links-legible-per-scheme @inferred @source:docs/stylesheets/extra.css#[data-md-color-scheme="slate"]
  Scenario Outline: Links use the accent that stays readable in the chosen colour scheme
    Given the reader views the docs in the <scheme> colour scheme
    When the reader reads a page containing links
    Then links are shown in <link_colour>
    And the page background is <background>

    Examples:
      | scheme | link_colour   | background                                                       |
      | light  | teal #0ea5a5  | the theme's default light background                             |
      | dark   | green #34d399 | near-black #0b0f1a, with the header and tabs on the same #0b0f1a |

  @id:docs.documentation-site.branding.003 @rule:docs.branding.no-borrowed-logo @inferred @source:docs/stylesheets/extra.css#.md-header__button.md-logo
  Scenario: The docs header shows no stand-in logo while the project has none
    Given the project has not configured a logo for the docs
    When a reader opens any docs page
    Then the header shows the site name without a logo icon
    And the navigation drawer title shows no logo icon
