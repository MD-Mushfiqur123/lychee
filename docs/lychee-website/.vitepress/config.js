import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "Lychee",
  description: "The fastest, most capable local LLM engine.",
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/getting-started' },
      { text: 'GitHub', link: 'https://github.com/MD-Mushfiqur123/lychee' }
    ],
    sidebar: [
      {
        text: 'Introduction',
        items: [
          { text: 'Getting Started', link: '/getting-started' },
          { text: 'Migrating from Ollama', link: '/migration-from-ollama' },
          { text: 'Lychee vs Ollama', link: '/vs-ollama' }
        ]
      },
      {
        text: 'Features',
        items: [
          { text: 'Anthropic API', link: '/anthropic-api' },
          { text: 'Agent Mode', link: '/agent-mode' },
          { text: 'Image Generation', link: '/image-generation' }
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'CLI Commands', link: '/cli-reference' },
          { text: 'Environment Variables', link: '/environment-variables' },
          { text: 'Benchmarks', link: '/benchmarks' }
        ]
      }
    ]
  }
})
