# SitePod Examples

This directory contains example projects for testing SitePod deployments.

## simple-site

A basic static site with HTML, CSS, and JavaScript.

### Files

```
simple-site/
├── sitepod.toml         # SitePod configuration
└── dist/
    ├── index.html       # Main page
    └── assets/
        ├── style.css    # Styles
        └── app.js       # JavaScript
```

### Deploy

```bash
cd examples/simple-site

# Deploy to beta
sitepod deploy

# Deploy to production
sitepod deploy --prod

# Create preview
sitepod preview
```

### Testing Locally

1. Start the SitePod server:

```bash
cd server
go run . serve --http=:8080 --data=../data
```

2. Deploy the example:

```bash
cd examples/simple-site
sitepod deploy
```

3. Access the site at `http://demo-site.localhost:8080` (requires hosts file entry or local DNS).
