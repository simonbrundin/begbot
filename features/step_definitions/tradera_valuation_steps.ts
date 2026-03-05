import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface TraderaAPIResponse {
  averagePrice: number;
  lowestPrice: number;
  highestPrice: number;
  count: number;
}

interface ValuationResult {
  value: number;
  confidence: number;
}

function calculateConfidenceFromCount(count: number): number {
  if (count === 0) return 0;
  if (count < 10) return 0.3;
  if (count < 50) return 0.5;
  if (count < 100) return 0.7;
  return 0.84 + Math.min((count - 100) / 1000, 0.11); // up to 0.95
}

function valuateFromTraderaAPI(response: TraderaAPIResponse): { result: ValuationResult | null; error: Error | null } {
  if (response.averagePrice === 0 && response.count === 0) {
    return { result: null, error: new Error('no prices found on Tradera') };
  }
  const confidence = calculateConfidenceFromCount(response.count);
  return { result: { value: response.averagePrice, confidence }, error: null };
}

let traderaEnabled: boolean = false;
let mockAPIResponse: TraderaAPIResponse;
let valuationResult: ValuationResult | null;
let valuationError: Error | null;

Before(function () {
  traderaEnabled = false;
  mockAPIResponse = { averagePrice: 0, lowestPrice: 0, highestPrice: 0, count: 0 };
  valuationResult = null;
  valuationError = null;
});

Given('Tradera valuation is enabled', function () {
  traderaEnabled = true;
});

Given('Tradera valuation is disabled', function () {
  traderaEnabled = false;
});

Given('the Tradera API returns average price {int} SEK with {int} sold items', function (avgPrice: number, count: number) {
  mockAPIResponse = { averagePrice: avgPrice, lowestPrice: 160, highestPrice: 2900, count };
});

Given('the Tradera API returns zero average price and zero items', function () {
  mockAPIResponse = { averagePrice: 0, lowestPrice: 0, highestPrice: 0, count: 0 };
});

When('I valuate a product on Tradera', function () {
  if (!traderaEnabled) {
    valuationResult = null;
    valuationError = null;
    return;
  }
  const result = valuateFromTraderaAPI(mockAPIResponse);
  valuationResult = result.result;
  valuationError = result.error;
});

Then('the valuation value should be {int}', function (expected: number) {
  assert(valuationResult !== null, 'Expected non-nil valuation');
  assert.strictEqual(valuationResult!.value, expected);
});

Then('the confidence should be at least {float} for {int} items', function (minConf: number, _count: number) {
  assert(valuationResult !== null, 'Expected non-nil valuation');
  assert(valuationResult!.confidence >= minConf,
    `Expected confidence >= ${minConf}, got ${valuationResult!.confidence}`);
});

Then('the Tradera valuation should be nil', function () {
  assert(valuationResult === null, `Expected nil valuation, got ${JSON.stringify(valuationResult)}`);
});

Then('no Tradera error should be returned', function () {
  assert(valuationError === null, `Expected no error, got: ${valuationError?.message}`);
});

Then('a Tradera error should be returned', function () {
  assert(valuationError !== null, 'Expected an error but got none');
});
