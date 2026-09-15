# Documentation site

Hugo renders this directory and the repository's `docs/` directory into the static site.

Build the site locally with:

```sh
hugo --source website --gc --minify
```

The generated `website/public/` directory is disposable and must not be committed.

## GitHub Pages setup

In the repository's Pages settings, select **Deploy from a branch**, then select the `gh-pages` branch and the repository root.

Create a fine-grained personal access token or machine-user token with `contents: write` permission for this repository and save it as the `PAGES_DEPLOY_TOKEN` Actions secret.

The token is required because a `GITHUB_TOKEN` push does not trigger a GitHub Pages build from a branch.

The documentation workflow builds the site after changes to `main` and pushes only the generated output to `gh-pages`.
