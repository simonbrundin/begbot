import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

function formatCurrency(cents: number): string {
  return `${(cents / 100).toFixed(2)} kr`;
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('sv-SE');
}

interface TradeItem {
  buy_price: number;
  buy_shipping_cost: number;
  sell_price: number | null;
  sell_packaging_cost: number | null;
  sell_postage_cost: number | null;
  sell_shipping_collected: number | null;
}

function calculateProfit(item: TradeItem): number {
  const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0);
  const buyTotal = item.buy_price + item.buy_shipping_cost;
  const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0);
  return sellTotal - (buyTotal + sellCost);
}

let amountOre: number;
let currencyResult: string;
let dateStr: string;
let dateResult: string;
let tradeItem: TradeItem;
let profitResult: number;

Before(function () {
  amountOre = 0;
  currencyResult = '';
  dateStr = '';
  dateResult = '';
  tradeItem = { buy_price: 0, buy_shipping_cost: 0, sell_price: null, sell_packaging_cost: null, sell_postage_cost: null, sell_shipping_collected: null };
  profitResult = 0;
});

Given('the amount {int} öre', function (amount: number) {
  amountOre = amount;
});

When('I format it as currency', function () {
  currencyResult = formatCurrency(amountOre);
});

Then('the result should be {string}', function (expected: string) {
  assert.strictEqual(currencyResult || dateResult, expected);
});

Given('the date {string}', function (date: string) {
  dateStr = date;
});

When('I format it as a date', function () {
  dateResult = formatDate(dateStr);
});

Given('a trade item with:', function (table: any) {
  tradeItem = { buy_price: 0, buy_shipping_cost: 0, sell_price: null, sell_packaging_cost: null, sell_postage_cost: null, sell_shipping_collected: null };
  for (const row of table.rows()) {
    const [field, value] = row;
    const numValue = value === 'null' ? null : parseInt(value);
    (tradeItem as any)[field] = numValue;
  }
});

Given('a trade item with buy_price {int} and no sell price', function (buyPrice: number) {
  tradeItem = { buy_price: buyPrice, buy_shipping_cost: 0, sell_price: null, sell_packaging_cost: null, sell_postage_cost: null, sell_shipping_collected: null };
});

When('I calculate the profit', function () {
  profitResult = calculateProfit(tradeItem);
});

Then('the profit should be {int}', function (expected: number) {
  assert.strictEqual(profitResult, expected, `Expected profit ${expected}, got ${profitResult}`);
});
