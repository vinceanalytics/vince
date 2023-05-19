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
        collapsed: true,
        items: [
          { text: 'Installation', link: '/guide/install-vince' },
          { text: 'Cli', link: '/guide/cli-vince' },
        ]
      },
      {
        text: "Site Settings",
        collapsed: true,
        items: [
          {
            text: "Site Landing Page",
          },
          {
            text: "Site Setting",
          },
          {
            text: "Change Domain Name",
          },
          {
            text: "Invite team members,assign roles and remove users",
          },
          {
            text: "Open site to the public",
          },
          {
            text: "Share your stats with a private and secure link",
          },
          {
            text: "Send reports via email",
          },
          {
            text: "Alerts",
            items: [],
          },
          {
            text: "Exclude pages from being tracked",
          },
          {
            text: "Transfer ownership of a site",
          },
          {
            text: "Reset your site data",
          },
          {
            text: "Delete your site data and stats",
          },
        ],
      },
      {
        text: "Stats Dashboard",
        collapsed: true,
        items: [],
      },
      {
        text: "Goals and Custom EVents",
        collapsed: true,
        items: [],
      },
      {
        text: "API",
        collapsed: true,
        items: [
          {
            text: "Stats API reference"
          },
          {
            text: "Events API reference"
          },
          {
            text: "Site Provisioning API reference"
          },
        ],
      },
      {
        text: "Account Settings",
        collapsed: true,
        items: [
          {
            text: "Change your account email address"
          },
          {
            text: "Reset your account password"
          },
          {
            text: "Delete your account"
          },
        ],
      },

      {
        text: 'Kubernetes',
        collapsed: true,
        items: [
          {
            text: 'Installation', items: [
              {
                text: 'Out of cluster',
                link: '/guide/install-v8s'
              },
              {
                text: 'Using helm',
              },
            ]
          },
          { text: 'Cli', link: '/guide/cli-v8s' },
        ]
      },
      {
        text: "Contribute",
        collapsed: true,
        items: [],
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
