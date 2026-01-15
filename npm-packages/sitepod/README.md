# SitePod CLI

Self-hosted static site deployments - deploy once, rollback in seconds.

## Installation

```bash
npm install -g sitepod
```

## Quick Start

```bash
# Login to your SitePod server
sitepod login --endpoint https://your-server.com

# Deploy your static site
cd your-site
sitepod deploy
```

## Commands

- `sitepod deploy` - Deploy your site (auto login + init if needed)
- `sitepod deploy --prod` - Deploy to production
- `sitepod login` - Authenticate with your server
- `sitepod init` - Initialize project config
- `sitepod preview` - Create preview deployment
- `sitepod rollback` - Rollback to previous version
- `sitepod history` - View deployment history
- `sitepod domain` - Manage custom domains

## Documentation

Visit [sitepod.dev](https://sitepod.dev) for full documentation.

## License

MIT
