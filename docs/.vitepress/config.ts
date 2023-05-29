import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  description: "The Cloud Native Web Analytics Platform",
  titleTemplate: true,
  title: 'vince',
  ignoreDeadLinks: 'localhostLinks',
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
          { text: 'Usage', link: '/guide/usage' },
          { text: 'Adding your first site', link: '/guide/adding-first-website' },
          { text: 'Cli', link: '/guide/cli-vince' },
        ]
      },
      {
        text: "Site Settings",
        items: [
          {
            text: "Site Landing Page",
            link: "/guide/site-landing-page",
          },
          {
            text: "Site Setting",
            link: "/guide/site-setting",
          },
          {
            text: "Change Domain Name",
            link: "/guide/change-domain-name",
          },
          {
            text: "Invite team members,assign roles and remove users",
            link: "/guide/invite-team-members-assign-roles-and-remove-users",
          },
          {
            text: "Open site to the public",
            link: "/guide/open-site-to-the-public",
          },
          {
            text: "Share your stats with a private and secure link",
            link: "/guide/share-your-stats-with-a-private-and-secure-link",
          },
          {
            text: "Send reports via email",
            link: "/guide/send-reports-via-email",
          },
          {
            text: "Exclude pages from being tracked",
            link: "/guide/exclude-pages-from-being-tracked",
          },
          {
            text: "Transfer ownership of a site",
            link: "/guide/transfer-ownership-of-a-site",
          },
          {
            text: "Reset your site data",
            link: "/guide/reset-your-site-data",
          },
          {
            text: "Delete your site data and stats",
            link: "/guide/delete-your-site-data-and-stats",
          },
        ],
      },
      {
        text: "Alerts",
        items: [],
      },
      {
        text: "Stats Dashboard",
        items: [],
      },
      {
        text: "Goals and Custom Events",
        items: [],
      },
      {
        text: "API",
        items: [
          {
            text: "Stats API reference",
            link: "/guide/stats-api-reference",
          },
          {
            text: "Events API reference",
            link: "/guide/events-api-reference",
          },
          {
            text: "Site Provisioning API reference",
            link: "/guide/site-provisioning-api-reference",
          },
        ],
      },
      {
        text: "Account Settings",
        items: [
          {
            text: "Change your account email address",
            link: "/guide/change-your-account-email-address",
          },
          {
            text: "Reset your account password",
            link: "/guide/reset-your-account-password",
          },
          {
            text: "Delete your account",
            link: "/guide/delete-your-account",
          },
        ],
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
