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
  return 0.84 + Math.min((count - 100) / 1000, 0.11);
}

function valuateFromTraderaAPI(response: TraderaAPIResponse): { result: ValuationResult | null; error: Error | null } {
  if (response.averagePrice === 0 && response.count === 0) {
    return { result: null, error: new Error('inga priser hittades på Tradera') };
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

Given('Tradera-värdering är aktiverad', function () {
  traderaEnabled = true;
});

Given('Tradera-värdering är inaktiverad', function () {
  traderaEnabled = false;
});

Given('Tradera-API:et returnerar ett genomsnittspris på {int} SEK med {int} sålda artiklar', function (avgPrice: number, count: number) {
  mockAPIResponse = { averagePrice: avgPrice, lowestPrice: 160, highestPrice: 2900, count };
});

Given('Tradera-API:et returnerar noll genomsnittspris och noll artiklar', function () {
  mockAPIResponse = { averagePrice: 0, lowestPrice: 0, highestPrice: 0, count: 0 };
});

When('jag värderar en produkt på Tradera', function () {
  if (!traderaEnabled) { valuationResult = null; valuationError = null; return; }
  const result = valuateFromTraderaAPI(mockAPIResponse);
  valuationResult = result.result;
  valuationError = result.error;
});

Then('värderingsvärdet ska vara {int}', function (expected: number) {
  assert(valuationResult !== null, 'Förväntade icke-null värdering');
  assert.strictEqual(valuationResult!.value, expected);
});

Then('förtroendet ska vara minst {float} för {int} artiklar', function (minConf: number, _count: number) {
  assert(valuationResult !== null, 'Förväntade icke-null värdering');
  assert(valuationResult!.confidence >= minConf,
    `Förväntade förtroende >= ${minConf}, fick ${valuationResult!.confidence}`);
});

Then('Tradera-värderingen ska vara null', function () {
  assert(valuationResult === null, `Förväntade null värdering, fick ${JSON.stringify(valuationResult)}`);
});

Then('inget Tradera-fel ska returneras', function () {
  assert(valuationError === null, `Förväntade inget fel, fick: ${valuationError?.message}`);
});

Then('ett Tradera-fel ska returneras', function () {
  assert(valuationError !== null, 'Förväntade ett fel men fick inget');
});
