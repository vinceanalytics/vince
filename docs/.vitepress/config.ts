import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  description: "The Cloud Native Web Analytics Platform",
  titleTemplate: false,
  title: '+',
  head: [
    ["meta", { name: "msapplication-TileColor", content: "#bdfcff" }],
    ["meta", { name: "twitter:card", content: "product" }],
    ["meta", { name: "twitter:site", content: "@gernesti" }],
    ["meta", { name: "twitter:title", content: "Vince analytics" }],
    ["meta", { name: "twitter:description", content: "The Cloud Native Web Analytics Platform." }],
    ["meta", { name: "og:url", content: "https://vinceanalytics.com" }],
    ["meta", { name: "og:title", content: "Vince Analytics" }],
    ["meta", { name: "og:description", content: "The Cloud Native Web Analytics Platform." }],
    ["meta", { name: "og:type", content: "article" }],
  ],
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
          { text: 'Installation', link: '/guide/install-vince' },
          { text: 'Cli', link: '/guide/cli-vince' },
        ]
      },
      {
        text: 'Kubernetes',
        items: [
          {
            text: 'Installation', items: [
              {
                text: 'Out of cluster',
                link: '/guide/install-v8s'
              },
            ]
          },
          { text: 'Cli', link: '/guide/cli-v8s' },
        ]
      },
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
