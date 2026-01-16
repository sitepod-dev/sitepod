---
title: CI/CD Integration
description: Automate deployments with GitHub Actions, GitLab CI, and more
---

Integrate SitePod into your CI/CD pipeline for automatic deployments.

## API tokens

First, create an API token for CI/CD use:

```bash
sitepod token create --name "github-actions"
```

```
✓ Created API Token

  Token: sitepod_xxxxxxxxxxxxxxxxxxxx

  Store this token securely. It won't be shown again.
  Add to GitHub Secrets as: SITEPOD_TOKEN
```

## GitHub Actions

### Basic deployment

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Deploy to SitePod
        run: |
          curl -fsSL https://get.sitepod.dev | sh
          sitepod deploy --prod
        env:
          SITEPOD_TOKEN: ${{ secrets.SITEPOD_TOKEN }}
```

### Preview deployments on PR

```yaml
# .github/workflows/preview.yml
name: Preview

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install and build
        run: |
          npm ci
          npm run build

      - name: Create preview
        id: preview
        run: |
          curl -fsSL https://get.sitepod.dev | sh
          sitepod preview --slug pr-${{ github.event.pull_request.number }}
          echo "url=https://my-site--pr-${{ github.event.pull_request.number }}.preview.sitepod.dev" >> $GITHUB_OUTPUT
        env:
          SITEPOD_TOKEN: ${{ secrets.SITEPOD_TOKEN }}

      - name: Comment on PR
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '### Preview deployed\n\n${{ steps.preview.outputs.url }}'
            })
```

### Using custom server

```yaml
- name: Deploy to SitePod
  run: |
    curl -fsSL https://get.sitepod.dev | sh
    sitepod deploy --prod
  env:
    SITEPOD_TOKEN: ${{ secrets.SITEPOD_TOKEN }}
    SITEPOD_ENDPOINT: https://sitepod.example.com
```

## GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - build
  - deploy

build:
  stage: build
  image: node:20
  script:
    - npm ci
    - npm run build
  artifacts:
    paths:
      - dist/

deploy:
  stage: deploy
  image: ubuntu:latest
  script:
    - curl -fsSL https://get.sitepod.dev | sh
    - sitepod deploy --prod
  variables:
    SITEPOD_TOKEN: $SITEPOD_TOKEN
  only:
    - main
```

## CircleCI

```yaml
# .circleci/config.yml
version: 2.1

jobs:
  build-and-deploy:
    docker:
      - image: cimg/node:20.0
    steps:
      - checkout
      - run:
          name: Install dependencies
          command: npm ci
      - run:
          name: Build
          command: npm run build
      - run:
          name: Deploy
          command: |
            curl -fsSL https://get.sitepod.dev | sh
            sitepod deploy --prod

workflows:
  deploy:
    jobs:
      - build-and-deploy:
          filters:
            branches:
              only: main
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `SITEPOD_TOKEN` | API token for authentication |
| `SITEPOD_ENDPOINT` | Server URL (optional, defaults to sitepod.dev) |

## Best practices

### 1. Use secrets

Never commit tokens to your repository. Use your CI provider's secret management:
- GitHub: Repository Settings → Secrets
- GitLab: Settings → CI/CD → Variables
- CircleCI: Project Settings → Environment Variables

### 2. Separate staging and production

```yaml
deploy-staging:
  if: github.ref == 'refs/heads/develop'
  run: sitepod deploy  # deploys to beta

deploy-production:
  if: github.ref == 'refs/heads/main'
  run: sitepod deploy --prod
```

### 3. Add deployment message

Include git info in deployments:

```yaml
- name: Deploy
  run: |
    sitepod deploy --prod --message "$(git log -1 --pretty=%B)"
```

### 4. Cache CLI installation

```yaml
- name: Cache SitePod CLI
  uses: actions/cache@v4
  with:
    path: ~/.local/bin/sitepod
    key: sitepod-cli-${{ runner.os }}

- name: Install SitePod CLI
  run: |
    if [ ! -f ~/.local/bin/sitepod ]; then
      curl -fsSL https://get.sitepod.dev | sh
    fi
```

## See also

- [Preview Deployments](/docs/guides/previews/) - Temporary preview URLs
- [API Reference](/docs/api/overview/) - Direct API integration
