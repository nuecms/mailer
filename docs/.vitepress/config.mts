import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Go Mail Server',
  description: '轻量级 SMTP 邮件发送服务',
  base: '/',
  lastUpdated: true,
  
  head: [
    ['link', { rel: 'icon', href: '/images/logo.svg' }]
  ],
  
  themeConfig: {
    logo: '/images/logo.svg',
    
    nav: [
      { text: '首页', link: '/' },
      { text: '指南', link: '/guides/deployment' },
      { text: '示例', link: '/examples/python' },
      { text: 'API', link: '/api/overview' },
      { text: 'GitHub', link: 'https://github.com/nuecms/mailer' }
    ],
    
    sidebar: {
      '/guides/': [
        {
          text: '使用指南',
          items: [
            { text: '部署指南', link: '/guides/deployment' },
            { text: '配置详解', link: '/guides/configuration' },
            { text: '直接发送模式', link: '/guides/direct-delivery' },
            { text: '高级功能', link: '/guides/advanced-features' },
            { text: 'DKIM 设置', link: '/guides/dkim-setup' },
            { text: '多提供商故障转移', link: '/guides/provider_failover' },
            { text: '性能优化', link: '/guides/optimization' },
            { text: '故障排查', link: '/guides/troubleshooting' }
          ]
        }
      ],
      '/examples/': [
        {
          text: '使用示例',
          items: [
            { text: 'Python', link: '/examples/python' },
            { text: 'Node.js', link: '/examples/nodejs' },
            { text: 'Go', link: '/examples/go' },
            { text: 'PHP', link: '/examples/php' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'API 参考',
          items: [
            { text: '概述', link: '/api/overview' },
            { text: '健康检查', link: '/api/health' },
            { text: '指标', link: '/api/metrics' },
            { text: '管理操作', link: '/api/admin' }
          ]
        }
      ]
    },
    
    socialLinks: [
      { icon: 'github', link: 'https://github.com/nuecms/mailer' }
    ],
    
    footer: {
      message: '基于 MIT 许可证发布',
      copyright: 'Copyright © 2023-present NueCMS 团队'
    },
    
    search: {
      provider: 'local'
    },
    
    editLink: {
      pattern: 'https://github.com/nuecms/mailer/edit/main/docs/:path',
      text: '在 GitHub 上编辑此页'
    }
  }
})
