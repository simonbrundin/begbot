import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface ValuationTypeConfig {
  type: number;
  active: boolean;
  weight: number;
}

function normalizeWeights(configs: ValuationTypeConfig[]): ValuationTypeConfig[] {
  const result = configs.map(c => ({ ...c }));
  const active = result.filter(c => c.active);
  for (const c of result) { if (!c.active) c.weight = 0; }
  if (active.length === 0) { for (const c of result) c.weight = 0; return result; }
  const hasZeroWeight = active.some(c => c.weight === 0);
  if (hasZeroWeight) {
    const equalShare = 100 / active.length;
    for (const c of result) { if (c.active) c.weight = equalShare; }
    return result;
  }
  const totalWeight = active.reduce((sum, c) => sum + c.weight, 0);
  for (const c of result) { if (c.active) c.weight = (c.weight / totalWeight) * 100; }
  return result;
}

let inputConfigs: ValuationTypeConfig[];
let normalizedConfigs: ValuationTypeConfig[];
let originalWeights: number[];

Before(function () {
  inputConfigs = [];
  normalizedConfigs = [];
  originalWeights = [];
});

Given('{int} aktiva värderingstyper med vikten {int} var', function (count: number, weight: number) {
  inputConfigs = Array.from({ length: count }, (_, i) => ({ type: i + 1, active: true, weight }));
});

Given('värderingstyper:', function (table: any) {
  inputConfigs = [];
  for (const row of table.rows()) {
    inputConfigs.push({ type: parseInt(row[0]), active: row[1] === 'true', weight: parseFloat(row[2]) });
  }
});

Given('{int} aktiv värderingstyp med vikten {int}', function (_count: number, weight: number) {
  inputConfigs = [{ type: 1, active: true, weight }];
});

When('jag normaliserar vikterna', function () {
  originalWeights = inputConfigs.map(c => c.weight);
  normalizedConfigs = normalizeWeights(inputConfigs);
});

Then('ska varje aktiv typ ha vikten ungefär {float}', function (expected: number) {
  for (const c of normalizedConfigs) {
    if (c.active) {
      assert(Math.abs(c.weight - expected) < 0.01, `Förväntade vikt ~${expected}, fick ${c.weight}`);
    }
  }
});

Then('ska summan av aktiva vikter vara {int}', function (expected: number) {
  const sum = normalizedConfigs.filter(c => c.active).reduce((s, c) => s + c.weight, 0);
  assert(Math.abs(sum - expected) < 0.001, `Förväntade summa ${expected}, fick ${sum}`);
});

Then('den inaktiva typens vikt ska vara {int}', function (expected: number) {
  for (const c of normalizedConfigs) {
    if (!c.active) assert.strictEqual(c.weight, expected);
  }
});

Then('ska typ {int} ha vikten {int}', function (typeNum: number, expected: number) {
  const config = normalizedConfigs.find(c => c.type === typeNum);
  assert(config !== undefined, `Typ ${typeNum} hittades inte`);
  assert(Math.abs(config!.weight - expected) < 0.001,
    `Förväntade typ ${typeNum} vikt ${expected}, fick ${config!.weight}`);
});

Then('ska alla aktiva typer ha lika vikt ungefär {float}', function (expected: number) {
  for (const c of normalizedConfigs) {
    if (c.active) {
      assert(Math.abs(c.weight - expected) < 0.01, `Förväntade vikt ~${expected}, fick ${c.weight}`);
    }
  }
});

Then('ska alla typer ha vikten {int}', function (expected: number) {
  for (const c of normalizedConfigs) {
    assert.strictEqual(c.weight, expected, `Förväntade vikt ${expected}, fick ${c.weight}`);
  }
});

Then('ska den typen ha vikten {int}', function (expected: number) {
  assert(normalizedConfigs.length > 0, 'Inga konfigurationer');
  assert(Math.abs(normalizedConfigs[0].weight - expected) < 0.001,
    `Förväntade vikt ${expected}, fick ${normalizedConfigs[0].weight}`);
});

Then('ska de ursprungliga indata-vikterna vara oförändrade', function () {
  for (let i = 0; i < inputConfigs.length; i++) {
    assert.strictEqual(inputConfigs[i].weight, originalWeights[i],
      `Förväntade ursprunglig vikt ${originalWeights[i]} att vara oförändrad, fick ${inputConfigs[i].weight}`);
  }
});
