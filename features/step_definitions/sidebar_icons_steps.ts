import { Given, When, Then, Before } from '@cucumber/cucumber';
import { readFileSync } from 'fs';
import { resolve } from 'path';
import assert from 'assert';

const layoutPath = resolve('./frontend/layouts/default.vue');
const packageJsonPath = resolve('./frontend/package.json');
const nuxtConfigPath = resolve('./frontend/nuxt.config.ts');

let layoutContent: string = '';
let packageJson: Record<string, any> = {};
let nuxtConfigContent: string = '';

Before(function () {
  layoutContent = '';
  packageJson = {};
  nuxtConfigContent = '';
});

Given('the Nuxt frontend application', function () {
  // App context is established
});

When('I check the package.json', function () {
  try {
    packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf-8'));
  } catch {
    packageJson = {};
  }
});

When('I check nuxt.config.ts', function () {
  try {
    nuxtConfigContent = readFileSync(nuxtConfigPath, 'utf-8');
  } catch {
    nuxtConfigContent = '';
  }
});

When('I check the sidebar layout', function () {
  try {
    layoutContent = readFileSync(layoutPath, 'utf-8');
  } catch {
    layoutContent = '';
  }
});

Then('nuxt icon module should be in devDependencies', function () {
  assert(packageJson.devDependencies && packageJson.devDependencies['@nuxt/icon'],
    'Expected @nuxt/icon to be in devDependencies');
});

Then('nuxt config should include the icon module', function () {
  assert(nuxtConfigContent.includes('@nuxt/icon'),
    'Expected nuxt.config.ts to contain @nuxt/icon');
});

Then('lucide icon collection should be installed', function () {
  const hasLucide = packageJson.devDependencies?.['@iconify-json/lucide'];
  const hasAll = packageJson.devDependencies?.['@iconify/json'];
  assert(hasLucide || hasAll,
    'Expected @iconify-json/lucide or @iconify/json to be installed');
});

Then('it should contain a lucide:home icon', function () {
  assert(layoutContent.match(/lucide:home|lucide:Home/),
    'Expected layout to contain lucide:home icon');
});

Then('it should contain a lucide:package icon', function () {
  assert(layoutContent.match(/lucide:package|lucide:Package/),
    'Expected layout to contain lucide:package icon');
});

Then('it should contain a lucide:list icon', function () {
  assert(layoutContent.match(/lucide:list|lucide:List/),
    'Expected layout to contain lucide:list icon');
});

Then('it should contain a lucide:arrow-left-right icon', function () {
  assert(layoutContent.match(/lucide:arrow-left-right|lucide:ArrowLeftRight/),
    'Expected layout to contain lucide:arrow-left-right icon');
});

Then('it should contain a lucide:bar-chart icon', function () {
  assert(layoutContent.match(/lucide:bar-chart|lucide:BarChart/),
    'Expected layout to contain lucide:bar-chart icon');
});

Then('it should contain a lucide:spider icon', function () {
  assert(layoutContent.match(/lucide:spider|lucide:Spider/),
    'Expected layout to contain lucide:spider icon');
});

Then('it should contain a lucide:history icon', function () {
  assert(layoutContent.match(/lucide:history|lucide:History/),
    'Expected layout to contain lucide:history icon');
});

Then('it should contain a lucide:megaphone icon', function () {
  assert(layoutContent.match(/lucide:megaphone|lucide:Megaphone/),
    'Expected layout to contain lucide:megaphone icon');
});

Then('all Icon components should have consistent size styling', function () {
  const iconMatches = layoutContent.match(/<Icon[^>]*>/g) || [];
  if (iconMatches.length > 0) {
    const withSize = iconMatches.filter((i: string) => i.includes('size='));
    const withoutSize = iconMatches.filter((i: string) => !i.includes('size='));
    assert(withSize.length === 0 || withoutSize.length === 0,
      'Icons should have consistent size styling (all with size or all without)');
  }
});

Then('all Icon components should not have explicit color overrides', function () {
  const iconMatches = layoutContent.match(/<Icon[^>]*>/g) || [];
  for (const icon of iconMatches) {
    const hasNoColor = !icon.includes('color=') && !icon.includes('style=');
    const hasInheritColor = icon.includes('currentColor');
    assert(hasNoColor || hasInheritColor,
      `Icon should not have explicit color override: ${icon}`);
  }
});

Then('the layout should contain "hover:bg-slate-700"', function () {
  assert(layoutContent.includes('hover:bg-slate-700'),
    'Expected layout to contain hover:bg-slate-700');
});

Then('the layout should contain active-class styling', function () {
  assert(layoutContent.match(/active-class=("|')?bg-slate-700/),
    'Expected layout to contain active-class styling');
});

Then('there should be at least {int} NuxtLink elements with icons', function (minCount: number) {
  const nuxtLinkPattern = /<NuxtLink[^>]*>[\s\S]*?<Icon[^>]*>[\s\S]*?<\/NuxtLink>/g;
  const matches = layoutContent.match(nuxtLinkPattern) || [];
  assert(matches.length >= minCount,
    `Expected at least ${minCount} NuxtLink with icons, got ${matches.length}`);
});

Then('it should not use remote server bundle', function () {
  assert(!nuxtConfigContent.match(/serverBundle:\s*['"]remote['"]/),
    'Expected nuxt.config.ts to not use remote server bundle');
});
