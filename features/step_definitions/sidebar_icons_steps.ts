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

Given('Nuxt-frontendapplikationen', function () {});

When('jag kontrollerar package.json', function () {
  try { packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf-8')); }
  catch { packageJson = {}; }
});

When('jag kontrollerar nuxt.config.ts', function () {
  try { nuxtConfigContent = readFileSync(nuxtConfigPath, 'utf-8'); }
  catch { nuxtConfigContent = ''; }
});

When('jag kontrollerar sidofältslayouten', function () {
  try { layoutContent = readFileSync(layoutPath, 'utf-8'); }
  catch { layoutContent = ''; }
});

Then('Nuxt-ikonmodulen ska finnas i devDependencies', function () {
  assert(packageJson.devDependencies && packageJson.devDependencies['@nuxt/icon'],
    'Förväntade @nuxt/icon i devDependencies');
});

Then('Nuxt-konfigurationen ska inkludera ikonmodulen', function () {
  assert(nuxtConfigContent.includes('@nuxt/icon'), 'Förväntade nuxt.config.ts att innehålla @nuxt/icon');
});

Then('Lucide-ikonsamlingen ska vara installerad', function () {
  const hasLucide = packageJson.devDependencies?.['@iconify-json/lucide'];
  const hasAll = packageJson.devDependencies?.['@iconify/json'];
  assert(hasLucide || hasAll, 'Förväntade @iconify-json/lucide eller @iconify/json att vara installerad');
});

Then('ska den innehålla en lucide:home-ikon', function () {
  assert(layoutContent.match(/lucide:home|lucide:Home/), 'Förväntade layouten att innehålla lucide:home-ikon');
});

Then('ska den innehålla en lucide:package-ikon', function () {
  assert(layoutContent.match(/lucide:package|lucide:Package/), 'Förväntade layouten att innehålla lucide:package-ikon');
});

Then('ska den innehålla en lucide:list-ikon', function () {
  assert(layoutContent.match(/lucide:list|lucide:List/), 'Förväntade layouten att innehålla lucide:list-ikon');
});

Then('ska den innehålla en lucide:arrow-left-right-ikon', function () {
  assert(layoutContent.match(/lucide:arrow-left-right|lucide:ArrowLeftRight/), 'Förväntade layouten att innehålla lucide:arrow-left-right-ikon');
});

Then('ska den innehålla en lucide:bar-chart-ikon', function () {
  assert(layoutContent.match(/lucide:bar-chart|lucide:BarChart/), 'Förväntade layouten att innehålla lucide:bar-chart-ikon');
});

Then('ska den innehålla en lucide:spider-ikon', function () {
  assert(layoutContent.match(/lucide:spider|lucide:Spider/), 'Förväntade layouten att innehålla lucide:spider-ikon');
});

Then('ska den innehålla en lucide:history-ikon', function () {
  assert(layoutContent.match(/lucide:history|lucide:History/), 'Förväntade layouten att innehålla lucide:history-ikon');
});

Then('ska den innehålla en lucide:megaphone-ikon', function () {
  assert(layoutContent.match(/lucide:megaphone|lucide:Megaphone/), 'Förväntade layouten att innehålla lucide:megaphone-ikon');
});

Then('ska alla ikonkomponenter ha konsekvent storleksstil', function () {
  const iconMatches = layoutContent.match(/<Icon[^>]*>/g) || [];
  if (iconMatches.length > 0) {
    const withSize = iconMatches.filter((i: string) => i.includes('size='));
    const withoutSize = iconMatches.filter((i: string) => !i.includes('size='));
    assert(withSize.length === 0 || withoutSize.length === 0,
      'Ikoner ska ha konsekvent storleksstil');
  }
});

Then('ska inga ikonkomponenter ha explicita färgöverskridningar', function () {
  const iconMatches = layoutContent.match(/<Icon[^>]*>/g) || [];
  for (const icon of iconMatches) {
    const hasNoColor = !icon.includes('color=') && !icon.includes('style=');
    const hasInheritColor = icon.includes('currentColor');
    assert(hasNoColor || hasInheritColor, `Ikon ska inte ha explicit färgöverskridning: ${icon}`);
  }
});

Then('ska layouten innehålla {string}', function (expected: string) {
  assert(layoutContent.includes(expected), `Förväntade layouten att innehålla "${expected}"`);
});

Then('ska layouten innehålla active-class-stilsättning', function () {
  assert(layoutContent.match(/active-class=("|')?bg-slate-700/), 'Förväntade layouten att innehålla active-class-stilsättning');
});

Then('ska det finnas minst {int} NuxtLink-element med ikoner', function (minCount: number) {
  const nuxtLinkPattern = /<NuxtLink[^>]*>[\s\S]*?<Icon[^>]*>[\s\S]*?<\/NuxtLink>/g;
  const matches = layoutContent.match(nuxtLinkPattern) || [];
  assert(matches.length >= minCount, `Förväntade minst ${minCount} NuxtLink med ikoner, fick ${matches.length}`);
});

Then('ska konfigurationen inte använda fjärr-server-bundle', function () {
  assert(!nuxtConfigContent.match(/serverBundle:\s*['"]remote['"]/), 'Förväntade att nuxt.config.ts inte använder fjärr-server-bundle');
});
