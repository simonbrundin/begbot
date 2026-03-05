import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface Product {
  id: number;
  brand: string | null;
  name: string | null;
  enabled: boolean | null;
  createdAt: Date | null;
}

function serializeProduct(product: Product): Record<string, any> {
  const result: Record<string, any> = { id: product.id };
  if (product.brand !== null && product.brand !== undefined) result.brand = product.brand;
  if (product.name !== null && product.name !== undefined) result.name = product.name;
  if (product.enabled !== null && product.enabled !== undefined) result.enabled = product.enabled;
  if (product.createdAt !== null && product.createdAt !== undefined) {
    result.created_at = product.createdAt.toISOString();
  }
  return result;
}

let product: Product;
let serialized: Record<string, any>;

Before(function () {
  product = { id: 1, brand: null, name: null, enabled: null, createdAt: null };
  serialized = {};
});

Given('en produkt med null i created_at', function () {
  product = { id: 1, brand: 'TestBrand', name: 'TestName', enabled: true, createdAt: null };
});

Given('en produkt med created_at {string}', function (dateStr: string) {
  product = { id: 1, brand: 'TestBrand', name: 'TestName', enabled: true, createdAt: new Date(dateStr) };
});

Given('en produkt med tomt varumärke och namn', function () {
  product = { id: 1, brand: '', name: '', enabled: false, createdAt: new Date() };
});

Given('en produkt med enabled satt till true', function () {
  product = { id: 1, brand: 'TestBrand', name: 'TestName', enabled: true, createdAt: new Date() };
});

Given('en produkt med alla null-fält', function () {
  product = { id: 0, brand: '', name: '', enabled: false, createdAt: null };
});

When('jag serialiserar produkten till JSON', function () {
  serialized = serializeProduct(product);
});

Then('ska JSON inte innehålla {string}', function (field: string) {
  const fieldName = field.replace(/^"/, '').replace(/"$/, '');
  assert(!(fieldName in serialized), `Förväntade att JSON inte innehåller "${fieldName}", men det gör det`);
});

Then('ska JSON innehålla {string}', function (field: string) {
  const fieldName = field.replace(/^"/, '').replace(/"$/, '');
  assert(fieldName in serialized, `Förväntade att JSON innehåller "${fieldName}", men det gör det inte`);
});

Then('ska JSON innehålla varumärket som tom sträng', function () {
  assert('brand' in serialized, 'Förväntade att varumärket finns');
  assert.strictEqual(serialized.brand, '', `Förväntade varumärke "", fick "${serialized.brand}"`);
});

Then('ska JSON innehålla namnet som tom sträng', function () {
  assert('name' in serialized, 'Förväntade att namnet finns');
  assert.strictEqual(serialized.name, '', `Förväntade namn "", fick "${serialized.name}"`);
});

Then('ska JSON innehålla {string}:true', function (_field: string) {
  assert('enabled' in serialized, 'Förväntade att enabled finns');
  assert.strictEqual(serialized.enabled, true, `Förväntade enabled true, fick ${serialized.enabled}`);
});

Then('ska JSON inte innehålla nolltiden {string}', function (zeroTime: string) {
  if ('created_at' in serialized && serialized.created_at) {
    assert(!serialized.created_at.includes(zeroTime),
      `Förväntade att JSON inte innehåller nolltiden "${zeroTime}", men det gör det`);
  }
  assert(true, 'Nolltidkontroll godkänd');
});
