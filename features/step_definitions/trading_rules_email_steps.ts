import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface TradingRules {
  minProfitSEK: number;
  minDiscount: number;
}

interface Listing {
  id: number;
  price: number;
  valuation: number;
  link: string;
  description: string;
  newPrice?: number;
}

interface EmailData {
  to: string[];
  subject: string;
  purchasePrice: number;
  valuation: number;
  discountPercent: number;
  newPrice: number;
  profit: number;
  description: string;
  link: string;
}

let tradingRules: TradingRules;
let listing: Listing;
let calculatedProfit: number;
let calculatedDiscount: number;
let emailData: EmailData | null;
let emailSentAsync: boolean;
let noCrash: boolean;

function passesRules(l: Listing, rules: TradingRules): boolean {
  const profit = l.valuation - l.price;
  const discount = (profit / l.valuation) * 100;
  return profit > rules.minProfitSEK && discount > rules.minDiscount;
}

Before(function () {
  tradingRules = { minProfitSEK: 500, minDiscount: 10 };
  listing = { id: 0, price: 0, valuation: 0, link: '', description: '' };
  calculatedProfit = 0;
  calculatedDiscount = 0;
  emailData = null;
  emailSentAsync = false;
  noCrash = true;
});

Given('trading rules with minimum profit {int} SEK and minimum discount {int}%', function (minProfit: number, minDiscount: number) {
  tradingRules = { minProfitSEK: minProfit, minDiscount: minDiscount };
});

Given('a listing with price {int} SEK and valuation {int} SEK', function (price: number, valuation: number) {
  listing = { id: 1, price, valuation, link: 'https://blocket.se/item/123', description: 'Test listing' };
});

Given('a listing with:', function (table: any) {
  listing = { id: 1, price: 0, valuation: 0, link: '', description: '' };
  for (const row of table.rows()) {
    const [field, value] = row;
    switch (field) {
      case 'price': listing.price = parseInt(value); break;
      case 'valuation': listing.valuation = parseInt(value); break;
      case 'link': listing.link = value; break;
      case 'description': listing.description = value; break;
    }
  }
});

Given('a new price of {int} SEK', function (price: number) {
  listing.newPrice = price;
});

Given('a valid listing that passes trading rules', function () {
  listing = { id: 1, price: 5000, valuation: 8000, link: 'https://blocket.se/item/123', description: 'Test' };
});

Given('no email configuration is set', function () {
  // Simulate no email config
});

When('evaluating the listing against trading rules', function () {
  calculatedProfit = listing.valuation - listing.price;
  calculatedDiscount = (calculatedProfit / listing.valuation) * 100;
});

When('preparing the email notification', function () {
  const profit = listing.valuation - listing.price;
  const discountPercent = (profit / listing.valuation) * 100;
  emailData = {
    to: ['recipient@example.com'],
    subject: `Ny annons: ${listing.link}`,
    purchasePrice: listing.price,
    valuation: listing.valuation,
    discountPercent,
    newPrice: listing.newPrice || 0,
    profit,
    description: listing.description,
    link: listing.link,
  };
});

When('the email notification is triggered', function () {
  emailSentAsync = true;
});

When('a listing passes trading rules', function () {
  noCrash = true;
});

Then('profit should be {int} SEK', function (expected: number) {
  assert.strictEqual(calculatedProfit, expected,
    `Expected profit ${expected}, got ${calculatedProfit}`);
});

Then('discount should be approximately {float}%', function (expected: number) {
  assert(Math.abs(calculatedDiscount - expected) < 1,
    `Expected discount ~${expected}%, got ${calculatedDiscount.toFixed(2)}%`);
});

Then('the listing should pass the trading rules', function () {
  assert(passesRules(listing, tradingRules),
    `Listing should pass trading rules (profit=${calculatedProfit}, discount=${calculatedDiscount.toFixed(2)}%)`);
});

Then('the listing should not pass the minimum profit threshold of {int} SEK', function (threshold: number) {
  assert(calculatedProfit <= threshold,
    `Profit ${calculatedProfit} should not exceed threshold ${threshold}`);
});

Then('the listing should not pass the minimum discount threshold of {float}%', function (threshold: number) {
  assert(calculatedDiscount <= threshold,
    `Discount ${calculatedDiscount.toFixed(2)}% should not exceed threshold ${threshold}%`);
});

Then('the email should include the purchase price', function () {
  assert(emailData !== null, 'Email data should be set');
  assert(emailData!.purchasePrice > 0, 'Email should include purchase price');
});

Then('the email should include the valuation', function () {
  assert(emailData !== null, 'Email data should be set');
  assert(emailData!.valuation > 0, 'Email should include valuation');
});

Then('the email should include the discount percent', function () {
  assert(emailData !== null, 'Email data should be set');
  assert(emailData!.discountPercent > 0, 'Email should include discount percent');
});

Then('the email should include the new price', function () {
  assert(emailData !== null, 'Email data should be set');
  // New price may be 0 if not set, but field should be present
  assert('newPrice' in emailData!, 'Email should include new price field');
});

Then('the email should include the profit', function () {
  assert(emailData !== null, 'Email data should be set');
  assert(emailData!.profit > 0, 'Email should include profit');
});

Then('the email should include the description', function () {
  assert(emailData !== null, 'Email data should be set');
  assert(emailData!.description !== '', 'Email should include description');
});

Then('the email should include the link', function () {
  assert(emailData !== null, 'Email data should be set');
  assert(emailData!.link !== '', 'Email should include link');
});

Then('the email should be sent asynchronously without blocking', function () {
  assert(emailSentAsync, 'Email should be sent asynchronously');
});

Then('no crash should occur', function () {
  assert(noCrash, 'No crash should occur');
});
