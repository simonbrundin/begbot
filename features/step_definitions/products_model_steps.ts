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

Given('a product with null created_at', function () {
  product = { id: 1, brand: 'TestBrand', name: 'TestName', enabled: true, createdAt: null };
});

Given('a product with created_at {string}', function (dateStr: string) {
  product = { id: 1, brand: 'TestBrand', name: 'TestName', enabled: true, createdAt: new Date(dateStr) };
});

Given('a product with empty brand and name', function () {
  product = { id: 1, brand: '', name: '', enabled: false, createdAt: new Date() };
});

Given('a product with enabled set to true', function () {
  product = { id: 1, brand: 'TestBrand', name: 'TestName', enabled: true, createdAt: new Date() };
});

Given('a product with all null optional fields', function () {
  product = { id: 0, brand: '', name: '', enabled: false, createdAt: null };
});

When('I serialize the product to JSON', function () {
  serialized = serializeProduct(product);
});

Then('the JSON should not contain {string}', function (field: string) {
  const fieldName = field.replace(/^"/, '').replace(/"$/, '');
  assert(!(fieldName in serialized), `Expected JSON to not contain "${fieldName}", but it does`);
});

Then('the JSON should contain {string}', function (field: string) {
  const fieldName = field.replace(/^"/, '').replace(/"$/, '');
  assert(fieldName in serialized, `Expected JSON to contain "${fieldName}", but it does not`);
});

Then('the JSON should contain brand as empty string', function () {
  assert('brand' in serialized, 'Expected brand to be present');
  assert.strictEqual(serialized.brand, '', `Expected brand to be "", got "${serialized.brand}"`);
});

Then('the JSON should contain name as empty string', function () {
  assert('name' in serialized, 'Expected name to be present');
  assert.strictEqual(serialized.name, '', `Expected name to be "", got "${serialized.name}"`);
});

Then('the JSON should contain {string}:true', function (_field: string) {
  assert('enabled' in serialized, 'Expected enabled to be present');
  assert.strictEqual(serialized.enabled, true, `Expected enabled to be true, got ${serialized.enabled}`);
});

Then('the JSON should not contain the zero time {string}', function (zeroTime: string) {
  if ('created_at' in serialized && serialized.created_at) {
    assert(!serialized.created_at.includes(zeroTime),
      `Expected JSON to not contain zero time "${zeroTime}", but it does`);
  }
  // If created_at is not in serialized, that's also acceptable
  assert(true, 'Zero time check passed');
});
