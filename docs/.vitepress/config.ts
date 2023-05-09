import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  description: "The Cloud Native Web Analytics Platform",
  title: "- Cloud Native Web Analytics",
  themeConfig: {
    logo: '/logo.svg',
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'guide', link: '/guide/' },
      { text: 'k8s', link: '/k8s/' },
      { text: 'blog', link: '/blog/' }
    ],

    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Installation', link: '/getting-started/install' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/vinceanalytics/vince' }
    ],
    footer: {
      message: "Released under the AGPL-3.0 license",
      copyright: "Copyright @ 2023-present Geofrey Ernest"
    }
  }
})
