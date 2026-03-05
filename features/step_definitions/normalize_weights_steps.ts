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
  
  // Set inactive weights to 0
  for (const c of result) {
    if (!c.active) c.weight = 0;
  }
  
  if (active.length === 0) {
    for (const c of result) c.weight = 0;
    return result;
  }
  
  // If any active type has weight 0, redistribute equally
  const hasZeroWeight = active.some(c => c.weight === 0);
  if (hasZeroWeight) {
    const equalShare = 100 / active.length;
    for (const c of result) {
      if (c.active) c.weight = equalShare;
    }
    return result;
  }
  
  // Normalize existing weights to sum to 100
  const totalWeight = active.reduce((sum, c) => sum + c.weight, 0);
  for (const c of result) {
    if (c.active) c.weight = (c.weight / totalWeight) * 100;
  }
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

Given('{int} active valuation types with weight {int} each', function (count: number, weight: number) {
  inputConfigs = Array.from({ length: count }, (_, i) => ({ type: i + 1, active: true, weight }));
});

Given('valuation types:', function (table: any) {
  inputConfigs = [];
  for (const row of table.rows()) {
    inputConfigs.push({
      type: parseInt(row[0]),
      active: row[1] === 'true',
      weight: parseFloat(row[2]),
    });
  }
});

Given('{int} active valuation type with weight {int}', function (_count: number, weight: number) {
  inputConfigs = [{ type: 1, active: true, weight }];
});

When('I normalize the weights', function () {
  originalWeights = inputConfigs.map(c => c.weight);
  normalizedConfigs = normalizeWeights(inputConfigs);
});

Then('each active type should have weight approximately {float}', function (expected: number) {
  for (const c of normalizedConfigs) {
    if (c.active) {
      assert(Math.abs(c.weight - expected) < 0.01,
        `Expected weight ~${expected}, got ${c.weight}`);
    }
  }
});

Then('the sum of active weights should be {int}', function (expected: number) {
  const sum = normalizedConfigs.filter(c => c.active).reduce((s, c) => s + c.weight, 0);
  assert(Math.abs(sum - expected) < 0.001, `Expected sum ${expected}, got ${sum}`);
});

Then('the inactive type weight should be {int}', function (expected: number) {
  for (const c of normalizedConfigs) {
    if (!c.active) {
      assert.strictEqual(c.weight, expected);
    }
  }
});

Then('type {int} should have weight {int}', function (typeNum: number, expected: number) {
  const config = normalizedConfigs.find(c => c.type === typeNum);
  assert(config !== undefined, `Type ${typeNum} not found`);
  assert(Math.abs(config!.weight - expected) < 0.001,
    `Expected type ${typeNum} weight ${expected}, got ${config!.weight}`);
});

Then('all active types should have equal weight approximately {float}', function (expected: number) {
  for (const c of normalizedConfigs) {
    if (c.active) {
      assert(Math.abs(c.weight - expected) < 0.01,
        `Expected weight ~${expected}, got ${c.weight}`);
    }
  }
});

Then('all types should have weight {int}', function (expected: number) {
  for (const c of normalizedConfigs) {
    assert.strictEqual(c.weight, expected, `Expected weight ${expected}, got ${c.weight}`);
  }
});

Then('that type should have weight {int}', function (expected: number) {
  assert(normalizedConfigs.length > 0, 'No configs');
  assert(Math.abs(normalizedConfigs[0].weight - expected) < 0.001,
    `Expected weight ${expected}, got ${normalizedConfigs[0].weight}`);
});

Then('the original input weights should be unchanged', function () {
  for (let i = 0; i < inputConfigs.length; i++) {
    assert.strictEqual(inputConfigs[i].weight, originalWeights[i],
      `Expected original weight ${originalWeights[i]} to be unchanged, got ${inputConfigs[i].weight}`);
  }
});
