# Documentation site

Hugo renders this directory and the repository's `docs/` directory into the static site.

Build the site locally with:

```sh
hugo --source website --gc --minify
```

The generated `website/public/` directory is disposable and must not be committed.

## GitHub Pages setup

In the repository's Pages settings, select **GitHub Actions** as the build and deployment source.

No deployment secret or `gh-pages` branch is required.

The documentation workflow builds the site after changes to `main`, uploads the generated output as a Pages artifact, and deploys it with `actions/deploy-pages`.
