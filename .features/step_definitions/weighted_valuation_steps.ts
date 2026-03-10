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
  productId: number, enabledTypes: { id: number }[], valuationsByProduct: ValuationsByProduct,
  weights: Weights, configsByProduct: ConfigsByProduct = {}
): { average: number; safetyPercent: number } | null {
  const activeTypes = enabledTypes.filter(vt => isTypeActiveForProduct(productId, vt.id, configsByProduct));
  if (activeTypes.length === 0) return null;
  const vals = valuationsByProduct[productId] ?? [];
  const getVal = (typeId: number) => vals.find(v => v.valuation_type_id === typeId) ?? null;
  const entries = activeTypes.map(vt => {
    const v = getVal(vt.id);
    return v !== null ? { valuation: v.valuation, weight: weights[vt.id] ?? 1 } : null;
  }).filter((e): e is { valuation: number; weight: number } => e !== null);
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

Given('inga värderingskonfigurationer finns för produkt {int}', function (_productId: number) {
  configsByProduct = {};
});

Given('en tom konfigurationslista för produkt {int}', function (productId: number) {
  configsByProduct = { [productId]: [] };
});

Given('en konfiguration som bara inaktiverar typ {int} för produkt {int}', function (typeId: number, productId: number) {
  configsByProduct = { [productId]: [{ product_id: productId, valuation_type_id: typeId, is_active: false }] };
});

Given('en konfiguration som inaktiverar typ {int} för produkt {int}', function (typeId: number, productId: number) {
  configsByProduct = { [productId]: [{ product_id: productId, valuation_type_id: typeId, is_active: false }] };
});

Given('en konfiguration som aktiverar typ {int} för produkt {int}', function (typeId: number, productId: number) {
  configsByProduct = { [productId]: [{ product_id: productId, valuation_type_id: typeId, is_active: true }] };
});

When('jag kontrollerar om typ {int} är aktiv för produkt {int}', function (typeId: number, productId: number) {
  isActiveResult = isTypeActiveForProduct(productId, typeId, configsByProduct);
});

Then('ska den vara aktiv', function () {
  assert(isActiveResult, 'Förväntade att typen är aktiv');
});

Then('ska den vara inaktiv', function () {
  assert(!isActiveResult, 'Förväntade att typen är inaktiv');
});

Given('inga aktiverade värderingstyper', function () {
  enabledTypes = [];
});

Given('aktiverade typer: {int}', function (t1: number) {
  enabledTypes = [{ id: t1 }];
});

Given('aktiverade typer: {int}, {int}', function (t1: number, t2: number) {
  enabledTypes = [{ id: t1 }, { id: t2 }];
});

Given('aktiverade typer: {int}, {int}, {int}', function (t1: number, t2: number, t3: number) {
  enabledTypes = [{ id: t1 }, { id: t2 }, { id: t3 }];
});

Given('produkt {int} saknar värderingar', function (productId: number) {
  valuationsByProduct = { [productId]: [] };
});

Given('produkt {int} har värderingen {int} för typ {int} och {int} för typ {int}', function (productId: number, val1: number, type1: number, val2: number, type2: number) {
  valuationsByProduct = {
    [productId]: [
      { valuation_type_id: type1, valuation: val1 },
      { valuation_type_id: type2, valuation: val2 },
    ],
  };
});

Given('produkt {int} har bara värderingen {int} för typ {int}', function (productId: number, val: number, typeId: number) {
  valuationsByProduct = { [productId]: [{ valuation_type_id: typeId, valuation: val }] };
});

Given('produkt {int} har värderingen {int} för typ {int} med vikten {int}', function (productId: number, val: number, typeId: number, weight: number) {
  valuationsByProduct = { [productId]: [{ valuation_type_id: typeId, valuation: val }] };
  weights = { [typeId]: weight };
});

Given('båda typerna har vikten {int}', function (weight: number) {
  for (const t of enabledTypes) { weights[t.id] = weight; }
});

Given('typ {int} har vikten {int} och typ {int} har vikten {int}', function (t1: number, w1: number, t2: number, w2: number) {
  weights = { [t1]: w1, [t2]: w2 };
});

Given('typ {int} är inaktiverad för produkt {int}', function (typeId: number, productId: number) {
  if (!configsByProduct[productId]) configsByProduct[productId] = [];
  configsByProduct[productId].push({ product_id: productId, valuation_type_id: typeId, is_active: false });
});

Given('både typ {int} och typ {int} är inaktiverade för produkt {int}', function (t1: number, t2: number, productId: number) {
  configsByProduct = {
    [productId]: [
      { product_id: productId, valuation_type_id: t1, is_active: false },
      { product_id: productId, valuation_type_id: t2, is_active: false },
    ],
  };
});

Given('inga konfigurationer finns', function () {
  configsByProduct = {};
});

Given('typ {int} och typ {int} är inaktiverade för produkt {int}', function (t1: number, t2: number, productId: number) {
  configsByProduct = {
    [productId]: [
      { product_id: productId, valuation_type_id: t1, is_active: false },
      { product_id: productId, valuation_type_id: t2, is_active: false },
    ],
  };
});

Given('produkt {int} har värderingen {int} för typ {int}, {int} för typ {int}, {int} för typ {int}', function (productId: number, val1: number, t1: number, val2: number, t2: number, val3: number, t3: number) {
  valuationsByProduct = {
    [productId]: [
      { valuation_type_id: t1, valuation: val1 },
      { valuation_type_id: t2, valuation: val2 },
      { valuation_type_id: t3, valuation: val3 },
    ],
  };
});

When('jag beräknar viktad värdering för produkt {int}', function (productId: number) {
  computeResult = computeWeightedValuation(productId, enabledTypes, valuationsByProduct, weights, configsByProduct);
});

Then('ska resultatet vara null', function () {
  assert(computeResult === null, `Förväntade null, fick ${JSON.stringify(computeResult)}`);
});

Then('ska genomsnittet vara {int}', function (expected: number) {
  assert(computeResult !== null, 'Förväntade icke-null resultat');
  assert.strictEqual(computeResult!.average, expected, `Förväntade genomsnitt ${expected}, fick ${computeResult!.average}`);
});

Then('ska säkerhetsprocenten vara {int}', function (expected: number) {
  assert(computeResult !== null, 'Förväntade icke-null resultat');
  assert.strictEqual(computeResult!.safetyPercent, expected, `Förväntade säkerhet ${expected}, fick ${computeResult!.safetyPercent}`);
});

Then('ska säkerhetsprocenten vara lägre än {int}', function (max: number) {
  assert(computeResult !== null, 'Förväntade icke-null resultat');
  assert(computeResult!.safetyPercent < max, `Förväntade säkerhet < ${max}, fick ${computeResult!.safetyPercent}`);
});
