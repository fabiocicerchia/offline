@domain:documentation-site @area:content-styling @from:stylesheets/extra
Feature: Reading tables in doc pages
  Readers on any screen width can read every column of the reference tables in the docs.

  @id:docs.documentation-site.content-styling.001 @rule:docs.content.wide-tables-not-clipped @inferred @source:docs/stylesheets/extra.css#.md-typeset table:not([class]) @questionable
  Scenario: A wide table is shown as a bordered card whose overflowing columns are hidden rather than scrollable
    Given a docs page contains a plain table wider than the reader's screen
    When the reader views that page on a narrow screen
    Then the table is drawn as a card with rounded corners and a hairline border
    And column headings appear in small uppercase
    And columns beyond the card's edge are hidden by the card
