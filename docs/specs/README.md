# Specifications of offline / docs

What this code does, one atomic Gherkin scenario per behaviour, read from the code by code-atlas.
Tags: `@verified` a test asserts it · `@inferred` read from the code only · `@gap` an entry point no test covers · `@human` written or amended by a person, never rewritten.

## Documentation site
The published docs and landing page and how they look and read.

- **Branding and theme**: [Presenting the offline brand on the docs and splash page](documentation-site/branding/docs-branding.feature) (3)
- **Content styling**: [Reading tables in doc pages](documentation-site/content-styling/docs-content-styling.feature) (1)
- **Self-contained assets**: [Rendering the docs without fetching fonts from the network](documentation-site/offline-assets/docs-self-contained-assets.feature) (1)
