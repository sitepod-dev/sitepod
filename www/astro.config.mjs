// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

import tailwindcss from '@tailwindcss/vite';

// https://astro.build/config
export default defineConfig({
  site: 'https://www.sitepod.dev',

  integrations: [
      starlight({
          title: 'SitePod',
          tagline: 'Deploy once, rollback in seconds',
          						logo: {
          							light: './src/assets/logo.svg',
          							dark: './src/assets/logo-dark.svg',
          							replacesTitle: true,
          						},          social: [
              { icon: 'github', label: 'GitHub', href: 'https://github.com/sitepod-dev/sitepod' },
          ],
          editLink: {
              baseUrl: 'https://github.com/sitepod-dev/sitepod/edit/main/www/',
          },
          customCss: [
              './src/styles/global.css',
              './src/styles/custom.css',
          ],
          			head: [
          				{
          					tag: 'link',
          					attrs: {
          						rel: 'preconnect',
          						href: 'https://fonts.googleapis.com',
          					},
          				},
          				{
          					tag: 'link',
          					attrs: {
          						rel: 'preconnect',
          						href: 'https://fonts.gstatic.com',
          						crossorigin: '',
          					},
          				},
          				{
          					tag: 'link',
          					attrs: {
          						href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500;600&display=swap',
          						rel: 'stylesheet',
          					},
          				},
          				{
          					tag: 'meta',
          					attrs: {
          						name: 'theme-color',
          						content: '#10b981',
          					},
          				},              {
                  tag: 'script',
                  content: `
                      // Default to dark theme
                      if (!localStorage.getItem('starlight-theme')) {
                          localStorage.setItem('starlight-theme', 'dark');
                      }
                  `,
              },
          ],
          sidebar: [
              {
                  label: 'Getting Started',
                  items: [
                      { label: 'Introduction', slug: 'docs/introduction' },
                      { label: 'Quick Start', slug: 'docs/quickstart' },
                      { label: 'Core Concepts', slug: 'docs/concepts' },
                  ],
              },
              {
                  label: 'CLI Reference',
                  autogenerate: { directory: 'docs/cli' },
              },
              {
                  label: 'Self-Hosting',
                  items: [
                      { label: 'Overview', slug: 'docs/self-hosting/overview' },
                      { label: 'VPS Deployment', slug: 'docs/self-hosting/vps' },
                      { label: 'Docker Compose', slug: 'docs/self-hosting/docker-compose' },
                      { label: 'Kubernetes', slug: 'docs/self-hosting/kubernetes' },
                      { label: 'Behind Reverse Proxy', slug: 'docs/self-hosting/reverse-proxy' },
                      { label: 'SSL/TLS Options', slug: 'docs/self-hosting/ssl' },
                      { label: 'Storage Backends', slug: 'docs/self-hosting/storage' },
                  ],
              },
              {
                  label: 'Guides',
                  items: [
                      { label: 'Custom Domains', slug: 'docs/guides/custom-domains' },
                      { label: 'CI/CD Integration', slug: 'docs/guides/ci-cd' },
                      { label: 'Preview Deployments', slug: 'docs/guides/previews' },
                      { label: 'Rollback', slug: 'docs/guides/rollback' },
                  ],
              },
              {
                  label: 'API Reference',
                  autogenerate: { directory: 'docs/api' },
              },
          ],
          defaultLocale: 'root',
          locales: {
              root: {
                  label: 'English',
                  lang: 'en',
              },
              zh: {
                  label: '简体中文',
                  lang: 'zh-CN',
              },
          },
      }),
	],

  vite: {
    plugins: [tailwindcss()],
  },
});