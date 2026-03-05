import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

type ConfigsByProduct = Record<number, { product_id: number; valuation_type_id: number; is_active: boolean }[]>;
type ValuationsByProduct = Record<number, { valuation_type_id: number | null; valuation: number }[]>;
type Weights = Record<number, number>;

function isTypeActiveForProduct(productId: number, typeId: number, configsByProduct: ConfigsByProduct): boolean {
  const configs = configsByProduct[productId];
  if (!configs || configs.length === 0) return true;
  const config = configs.find(c => c.valuation_type_id === typeId);
  if (!config) return true;
  return config.is_active;
}

function computeWeightedValuation(
  productId: number,
  enabledTypes: { id: number }[],
  valuationsByProduct: ValuationsByProduct,
  weights: Weights,
  configsByProduct: ConfigsByProduct = {}
): { average: number; safetyPercent: number } | null {
  const activeTypes = enabledTypes.filter(vt => isTypeActiveForProduct(productId, vt.id, configsByProduct));
  if (activeTypes.length === 0) return null;

  const vals = valuationsByProduct[productId] ?? [];
  const getVal = (typeId: number) => vals.find(v => v.valuation_type_id === typeId) ?? null;

  const entries = activeTypes
    .map(vt => {
      const v = getVal(vt.id);
      return v !== null ? { valuation: v.valuation, weight: weights[vt.id] ?? 1 } : null;
    })
    .filter((e): e is { valuation: number; weight: number } => e !== null);

  if (entries.length === 0) return null;
  const totalWeight = entries.reduce((s, e) => s + e.weight, 0);
  if (totalWeight === 0) return null;
  const average = entries.reduce((s, e) => s + e.valuation * e.weight, 0) / totalWeight;
  let safetyPercent = 100;
  if (entries.length > 1) {
    const mean = entries.reduce((s, e) => s + e.valuation, 0) / entries.length;
    const variance = entries.reduce((s, e) => s + Math.pow(e.valuation - mean, 2), 0) / entries.length;
    const stdDev = Math.sqrt(variance);
    safetyPercent = mean !== 0 ? Math.max(0, Math.round(100 - (stdDev / Math.abs(mean) * 100))) : 0;
  }
  return { average: Math.round(average), safetyPercent };
}

let configsByProduct: ConfigsByProduct;
let valuationsByProduct: ValuationsByProduct;
let enabledTypes: { id: number }[];
let weights: Weights;
let isActiveResult: boolean;
let computeResult: { average: number; safetyPercent: number } | null;

Before(function () {
  configsByProduct = {};
  valuationsByProduct = {};
  enabledTypes = [];
  weights = {};
  isActiveResult = false;
  computeResult = null;
});

Given('no valuation configs exist for product {int}', function (_productId: number) {
  configsByProduct = {};
});

Given('an empty config list for product {int}', function (productId: number) {
  configsByProduct = { [productId]: [] };
});

Given('a config that only deactivates type {int} for product {int}', function (typeId: number, productId: number) {
  configsByProduct = { [productId]: [{ product_id: productId, valuation_type_id: typeId, is_active: false }] };
});

Given('a config that deactivates type {int} for product {int}', function (typeId: number, productId: number) {
  configsByProduct = { [productId]: [{ product_id: productId, valuation_type_id: typeId, is_active: false }] };
});

Given('a config that activates type {int} for product {int}', function (typeId: number, productId: number) {
  configsByProduct = { [productId]: [{ product_id: productId, valuation_type_id: typeId, is_active: true }] };
});

When('I check if type {int} is active for product {int}', function (typeId: number, productId: number) {
  isActiveResult = isTypeActiveForProduct(productId, typeId, configsByProduct);
});

Then('it should be active', function () {
  assert(isActiveResult, 'Expected type to be active');
});

Then('it should be inactive', function () {
  assert(!isActiveResult, 'Expected type to be inactive');
});

Given('no enabled valuation types', function () {
  enabledTypes = [];
});

Given('enabled types: {int}', function (t1: number) {
  enabledTypes = [{ id: t1 }];
});

Given('enabled types: {int}, {int}', function (t1: number, t2: number) {
  enabledTypes = [{ id: t1 }, { id: t2 }];
});

Given('enabled types: {int}, {int}, {int}', function (t1: number, t2: number, t3: number) {
  enabledTypes = [{ id: t1 }, { id: t2 }, { id: t3 }];
});

Given('product {int} has no valuations', function (productId: number) {
  valuationsByProduct = { [productId]: [] };
});

Given('product {int} has valuation {int} for type {int} and {int} for type {int}', function (productId: number, val1: number, type1: number, val2: number, type2: number) {
  valuationsByProduct = {
    [productId]: [
      { valuation_type_id: type1, valuation: val1 },
      { valuation_type_id: type2, valuation: val2 },
    ],
  };
});

Given('product {int} has only valuation {int} for type {int}', function (productId: number, val: number, typeId: number) {
  valuationsByProduct = { [productId]: [{ valuation_type_id: typeId, valuation: val }] };
});

Given('product {int} has valuation {int} for type {int} and {int} for type {int} with weight {int}', function (productId: number, val1: number, type1: number, val2: number, type2: number) {
  valuationsByProduct = {
    [productId]: [
      { valuation_type_id: type1, valuation: val1 },
      { valuation_type_id: type2, valuation: val2 },
    ],
  };
});

Given('product {int} has valuation {int} for type {int} with weight {int}', function (productId: number, val: number, typeId: number, weight: number) {
  valuationsByProduct = { [productId]: [{ valuation_type_id: typeId, valuation: val }] };
  weights = { [typeId]: weight };
});

Given('both types have weight {int}', function (weight: number) {
  for (const t of enabledTypes) {
    weights[t.id] = weight;
  }
});

Given('type {int} has weight {int} and type {int} has weight {int}', function (t1: number, w1: number, t2: number, w2: number) {
  weights = { [t1]: w1, [t2]: w2 };
});

Given('type {int} is deactivated for product {int}', function (typeId: number, productId: number) {
  if (!configsByProduct[productId]) configsByProduct[productId] = [];
  configsByProduct[productId].push({ product_id: productId, valuation_type_id: typeId, is_active: false });
});

Given('both type {int} and type {int} are deactivated for product {int}', function (t1: number, t2: number, productId: number) {
  configsByProduct = {
    [productId]: [
      { product_id: productId, valuation_type_id: t1, is_active: false },
      { product_id: productId, valuation_type_id: t2, is_active: false },
    ],
  };
});

Given('no configs exist', function () {
  configsByProduct = {};
});

Given('types {int} and {int} are deactivated for product {int}', function (t1: number, t2: number, productId: number) {
  if (!configsByProduct[productId]) configsByProduct[productId] = [];
  configsByProduct[productId] = [
    { product_id: productId, valuation_type_id: t1, is_active: false },
    { product_id: productId, valuation_type_id: t2, is_active: false },
  ];
});

Given('product {int} has valuation {int} for type {int}, {int} for type {int}, {int} for type {int}', function (productId: number, val1: number, t1: number, val2: number, t2: number, val3: number, t3: number) {
  valuationsByProduct = {
    [productId]: [
      { valuation_type_id: t1, valuation: val1 },
      { valuation_type_id: t2, valuation: val2 },
      { valuation_type_id: t3, valuation: val3 },
    ],
  };
});

When('I compute weighted valuation for product {int}', function (productId: number) {
  computeResult = computeWeightedValuation(productId, enabledTypes, valuationsByProduct, weights, configsByProduct);
});

Then('the result should be null', function () {
  assert(computeResult === null, `Expected null, got ${JSON.stringify(computeResult)}`);
});

Then('the average should be {int}', function (expected: number) {
  assert(computeResult !== null, 'Expected non-null result');
  assert.strictEqual(computeResult!.average, expected,
    `Expected average ${expected}, got ${computeResult!.average}`);
});

Then('the safety percent should be {int}', function (expected: number) {
  assert(computeResult !== null, 'Expected non-null result');
  assert.strictEqual(computeResult!.safetyPercent, expected,
    `Expected safety ${expected}, got ${computeResult!.safetyPercent}`);
});

Then('the safety percent should be less than {int}', function (max: number) {
  assert(computeResult !== null, 'Expected non-null result');
  assert(computeResult!.safetyPercent < max,
    `Expected safety < ${max}, got ${computeResult!.safetyPercent}`);
});
