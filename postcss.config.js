import tailwind from 'tailwindcss'
import tailwindTypography from '@tailwindcss/typography'

export default {
  plugins: [
    tailwind({
      content: ['./blog/.vitepress/theme/**/*.vue'],
      plugins: [tailwindTypography]
    })
  ]
}
