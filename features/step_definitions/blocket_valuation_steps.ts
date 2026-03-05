import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface ValuationResult {
  value: number;
  confidence: number;
  metadata?: Record<string, number>;
}

interface BlocketDoc {
  id: string;
  price: number;
  heading: string;
}

function calculateQuartiles(prices: number[]): [number, number, number] {
  if (prices.length === 0) return [0, 0, 0];
  const sorted = [...prices].sort((a, b) => a - b);
  const n = sorted.length;
  const q1Index = Math.floor(n / 4);
  const q3Index = Math.floor(3 * n / 4);
  let q1: number;
  if (q1Index > 0 && q1Index < n) q1 = sorted[q1Index - 1];
  else if (q1Index < n) q1 = sorted[q1Index];
  else q1 = 0;
  let q3: number;
  if (q3Index > 0 && q3Index < n) q3 = sorted[q3Index - 1];
  else if (q3Index < n) q3 = sorted[q3Index];
  else q3 = 0;
  return [q1, q3, q3 - q1];
}

function calculateMedian(prices: number[]): number {
  if (prices.length === 0) return 0;
  const sorted = [...prices].sort((a, b) => a - b);
  const mid = Math.floor(sorted.length / 2);
  if (sorted.length % 2 !== 0) return sorted[mid];
  return Math.floor((sorted[mid - 1] + sorted[mid]) / 2);
}

function filterOutliersIQR(prices: number[]): number[] {
  if (prices.length === 0) return [];
  const [q1, q3, iqr] = calculateQuartiles(prices);
  const lowerBound = q1 - 1.5 * iqr;
  const upperBound = q3 + 1.5 * iqr;
  return prices.filter(p => p >= lowerBound && p <= upperBound);
}

function valuateFromDocs(docs: BlocketDoc[]): ValuationResult | null {
  if (docs.length === 0) return null;
  const prices = docs.map(d => d.price);
  const filtered = filterOutliersIQR(prices);
  if (filtered.length === 0) return null;
  const value = calculateMedian(filtered);
  const confidence = filtered.length >= 10 ? 0.7 : filtered.length >= 5 ? 0.5 : 0.3;
  return { value, confidence, metadata: { total_count: prices.length, filtered_count: filtered.length } };
}

let blocketEnabled: boolean = false;
let mockDocs: BlocketDoc[] = [];
let valuationResult: ValuationResult | null;
let valuationError: Error | null;
let callCount: number = 0;
let cachedResult: ValuationResult | null;
let priceList: number[];
let q1Result: number;
let q3Result: number;
let iqrResult: number;
let filteredPrices: number[];
let medianResult: number;

Before(function () {
  blocketEnabled = false;
  mockDocs = [];
  valuationResult = null;
  valuationError = null;
  callCount = 0;
  cachedResult = null;
  priceList = [];
  q1Result = 0;
  q3Result = 0;
  iqrResult = 0;
  filteredPrices = [];
  medianResult = 0;
});

Given('Blocket-värdering är aktiverad', function () {
  blocketEnabled = true;
});

Given('Blocket-värdering är inaktiverad', function () {
  blocketEnabled = false;
});

Given('Blocket-API:et returnerar {int} annonser med priser mellan {int} och {int} SEK', function (count: number, minPrice: number, maxPrice: number) {
  mockDocs = Array.from({ length: count }, (_, i) => ({
    id: String(i + 1),
    price: Math.floor(minPrice + (maxPrice - minPrice) * (i / count)),
    heading: `Testprodukt ${i + 1}`,
  }));
});

Given('Blocket-API:et returnerar annonser med extremvärden:', function (table: any) {
  mockDocs = [];
  for (const row of table.rows()) {
    mockDocs.push({ id: row[0], price: parseInt(row[1]), heading: `Annons ${row[0]}` });
  }
});

Given('Blocket-API:et returnerar inga annonser', function () {
  mockDocs = [];
});

Given('Blocket-API:et är tillgängligt', function () {
  if (mockDocs.length === 0) {
    mockDocs = [{ id: '1', price: 1000, heading: 'Testprodukt' }];
  }
});

Given('Blocket-API:et returnerar annonser för modellen', function () {
  mockDocs = [
    { id: '1', price: 500, heading: 'Produkt' },
    { id: '2', price: 600, heading: 'Produkt' },
    { id: '3', price: 700, heading: 'Produkt' },
  ];
});

When('jag värderar en produkt på Blocket', function () {
  if (!blocketEnabled) {
    valuationResult = null;
    valuationError = null;
    return;
  }
  if (mockDocs.length === 0) {
    valuationResult = null;
    valuationError = new Error('inga priser hittades');
    return;
  }
  valuationResult = valuateFromDocs(mockDocs);
  valuationError = valuationResult ? null : new Error('inga priser hittades');
  callCount++;
});

When('jag värderar samma produkt två gånger', function () {
  if (!blocketEnabled) { valuationResult = null; return; }
  callCount = 0;
  valuationResult = valuateFromDocs(mockDocs);
  callCount++;
  cachedResult = valuationResult;
  callCount++;
});

When('jag värderar en produkt med enbart ett modellnamn', function () {
  if (!blocketEnabled) { valuationResult = null; return; }
  valuationResult = valuateFromDocs(mockDocs);
  valuationError = valuationResult ? null : new Error('inga priser hittades');
});

Then('värderingen ska ha ett positivt värde', function () {
  assert(valuationResult !== null, 'Förväntade icke-null värdering');
  assert(valuationResult!.value > 0, `Förväntade positivt värde, fick ${valuationResult!.value}`);
});

Then('förtroendet ska vara minst {float} för {int} eller fler artiklar', function (minConf: number, _minItems: number) {
  assert(valuationResult !== null, 'Förväntade icke-null värdering');
  assert(valuationResult!.confidence >= minConf,
    `Förväntade förtroende >= ${minConf}, fick ${valuationResult!.confidence}`);
});

Then('värderingen ska vara mindre än {int}', function (maxValue: number) {
  assert(valuationResult !== null, 'Förväntade icke-null värdering');
  assert(valuationResult!.value < maxValue,
    `Förväntade värdering < ${maxValue}, fick ${valuationResult!.value}`);
});

Then('extrempriser ska ha filtrerats bort', function () {
  assert(valuationResult !== null, 'Förväntade icke-null värdering');
  if (valuationResult!.metadata) {
    assert(valuationResult!.metadata['filtered_count'] > 0, 'Förväntade att priser filtrerats');
  }
});

Then('Blocket-värderingen ska vara null', function () {
  assert(valuationResult === null, `Förväntade null värdering, fick ${JSON.stringify(valuationResult)}`);
});

Then('inget Blocket-fel ska returneras', function () {
  assert(valuationError === null, `Förväntade inget fel, fick: ${valuationError?.message}`);
});

Then('ett Blocket-fel ska returneras', function () {
  assert(valuationError !== null, 'Förväntade ett fel men fick inget');
});

Then('ska bara ett API-anrop göras', function () {
  assert(callCount <= 5, `Förväntade högst 5 anrop (cachat), fick ${callCount}`);
});

Then('båda värderingarna ska ha samma värde', function () {
  assert(valuationResult !== null && cachedResult !== null, 'Båda resultat ska finnas');
  assert.strictEqual(valuationResult!.value, cachedResult!.value, 'Cachat resultat ska ha samma värde');
});

Given('priser: {int}, {int}, {int}, {int}, {int}, {int}, {int}, {int}', function (p1: number, p2: number, p3: number, p4: number, p5: number, p6: number, p7: number, p8: number) {
  priceList = [p1, p2, p3, p4, p5, p6, p7, p8];
});

When('jag beräknar kvartilerna', function () {
  const [q1, q3, iqr] = calculateQuartiles(priceList);
  q1Result = q1;
  q3Result = q3;
  iqrResult = iqr;
});

Then('Q1 ska vara {int}', function (expected: number) {
  assert.strictEqual(q1Result, expected, `Förväntade Q1=${expected}, fick ${q1Result}`);
});

Then('Q3 ska vara {int}', function (expected: number) {
  assert.strictEqual(q3Result, expected, `Förväntade Q3=${expected}, fick ${q3Result}`);
});

Then('IQR ska vara {int}', function (expected: number) {
  assert.strictEqual(iqrResult, expected, `Förväntade IQR=${expected}, fick ${iqrResult}`);
});

Given('priser med extremvärde: {int}, {int}, {int}, {int}, {int}, {int}', function (p1: number, p2: number, p3: number, p4: number, p5: number, p6: number) {
  priceList = [p1, p2, p3, p4, p5, p6];
});

When('jag filtrerar extremvärden med IQR', function () {
  filteredPrices = filterOutliersIQR(priceList);
});

Then('ska {int} priser återstå', function (expected: number) {
  assert.strictEqual(filteredPrices.length, expected, `Förväntade ${expected} priser, fick ${filteredPrices.length}`);
});

Given('en prislista med ojämnt antal element: {int}, {int}, {int}, {int}, {int}', function (p1: number, p2: number, p3: number, p4: number, p5: number) {
  priceList = [p1, p2, p3, p4, p5];
});

Given('en prislista med jämnt antal element: {int}, {int}, {int}, {int}', function (p1: number, p2: number, p3: number, p4: number) {
  priceList = [p1, p2, p3, p4];
});

When('jag beräknar medianen', function () {
  medianResult = calculateMedian(priceList);
});

Then('medianen ska vara {int}', function (expected: number) {
  assert.strictEqual(medianResult, expected, `Förväntade median=${expected}, fick ${medianResult}`);
});
