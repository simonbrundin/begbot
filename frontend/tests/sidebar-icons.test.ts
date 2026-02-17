import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('Sidebar Icons Feature (#5)', () => {
  const layoutPath = resolve('./layouts/default.vue')
  const packageJsonPath = resolve('./package.json')
  const nuxtConfigPath = resolve('./nuxt.config.ts')

  // Test 1: @nuxt/icon module is installed
  describe('Module Installation', () => {
    it('should have @nuxt/icon in package.json dependencies', () => {
      const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf-8'))
      
      expect(packageJson.devDependencies).toHaveProperty('@nuxt/icon')
    })

    it('should have @nuxt/icon configured in nuxt.config.ts', () => {
      const nuxtConfigContent = readFileSync(nuxtConfigPath, 'utf-8')
      
      expect(nuxtConfigContent).toContain('@nuxt/icon')
    })

    it('should have lucide icon collection installed', () => {
      const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf-8'))
      
      // Either @iconify-json/lucide or @iconify/json should be installed
      const hasLucide = packageJson.devDependencies?.['@iconify-json/lucide']
      const hasAllIcons = packageJson.devDependencies?.['@iconify/json']
      
      expect(hasLucide || hasAllIcons).toBeDefined()
    })
  })

  // Test 2: Each menu item has an icon from Lucide
  describe('Menu Icons', () => {
    it('should render icon for Översikt menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      // Should contain Icon component with lucide:home for Översikt
      expect(layoutContent).toMatch(/lucide:home|lucide:Home/)
    })

    it('should render icon for Produkter menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/lucide:package|lucide:Package/)
    })

    it('should render icon for Mina annonser menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/lucide:list|lucide:List/)
    })

    it('should render icon for Transaktioner menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/lucide:arrow-left-right|lucide:ArrowLeftRight/)
    })

    it('should render icon for Marknadsanalys menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/lucide:bar-chart|lucide:BarChart/)
    })

    it('should render icon for Scraping menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/lucide:spider|lucide:Spider/)
    })

    it('should render icon for Historik menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/lucide:history|lucide:History/)
    })

    it('should render icon for Hittade annonser menu item', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/lucide:megaphone|lucide:Megaphone/)
    })
  })

  // Test 3: Icons are consistently styled
  describe('Icon Styling', () => {
    it('should have consistent icon size styling', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      // All Icon components should have a size prop or use default
      const iconMatches = layoutContent.match(/<Icon[^>]*>/g) || []
      
      // If there are Icon components, they should either all have size or use default
      if (iconMatches.length > 0) {
        const withSize = iconMatches.filter(i => i.includes('size='))
        const withoutSize = iconMatches.filter(i => !i.includes('size='))
        
        // Either all should have size, or none (relying on default)
        // But not a mix - that would be inconsistent
        expect(withSize.length === 0 || withoutSize.length === 0).toBe(true)
      }
    })

    it('should have icons with consistent color styling (matching text)', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      const iconMatches = layoutContent.match(/<Icon[^>]*>/g) || []
      
      // Check that icons have styling that matches the text color
      iconMatches.forEach(icon => {
        const hasNoColor = !icon.includes('color=') && !icon.includes('style=')
        const hasInheritColor = icon.includes('currentColor')
        
        expect(hasNoColor || hasInheritColor).toBe(true)
      })
    })
  })

  // Test 4: Hover/active states work with icons
  describe('Hover and Active States', () => {
    it('should preserve hover styling with icons present', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/hover:bg-slate-700/)
    })

    it('should preserve active state styling with icons present', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      expect(layoutContent).toMatch(/active-class=("|')?bg-slate-700/)
    })

    it('should have icons inside NuxtLink with consistent structure', () => {
      const layoutContent = readFileSync(layoutPath, 'utf-8')
      
      // Icons should be inside NuxtLink elements
      const nuxtLinkPattern = /<NuxtLink[^>]*>[\s\S]*?<Icon[^>]*>[\s\S]*?<\/NuxtLink>/g
      const matches = layoutContent.match(nuxtLinkPattern) || []
      
      // Should have at least 8 menu items with icons
      expect(matches.length).toBeGreaterThanOrEqual(8)
    })
  })

  // Edge case: Performance - icons should not significantly impact load time
  describe('Performance', () => {
    it('should use local icon collection (not remote)', () => {
      const nuxtConfigContent = readFileSync(nuxtConfigPath, 'utf-8')
      
      // Should NOT set serverBundle to 'remote'
      expect(nuxtConfigContent).not.toMatch(/serverBundle:\s*['"]remote['"]/)
    })
  })
})
