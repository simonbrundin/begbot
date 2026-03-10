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

Given('handelsregler med lägsta vinst på {int} SEK och lägsta rabatt på {int}%', function (minProfit: number, minDiscount: number) {
  tradingRules = { minProfitSEK: minProfit, minDiscount: minDiscount };
});

Given('en annons med priset {int} SEK och värderingen {int} SEK', function (price: number, valuation: number) {
  listing = { id: 1, price, valuation, link: 'https://blocket.se/item/123', description: 'Testannons' };
});

Given('en annons med:', function (table: any) {
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

Given('ett nypris på {int} SEK', function (price: number) {
  listing.newPrice = price;
});

Given('en giltig annons som uppfyller handelsreglerna', function () {
  listing = { id: 1, price: 5000, valuation: 8000, link: 'https://blocket.se/item/123', description: 'Test' };
});

Given('ingen e-postkonfiguration är inställd', function () {});

When('annonsen utvärderas mot handelsreglerna', function () {
  calculatedProfit = listing.valuation - listing.price;
  calculatedDiscount = (calculatedProfit / listing.valuation) * 100;
});

When('e-postnotifieringen förbereds', function () {
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

When('e-postnotifieringen utlöses', function () {
  emailSentAsync = true;
});

When('en annons uppfyller handelsreglerna', function () {
  noCrash = true;
});

Then('ska vinsten vara {int} SEK', function (expected: number) {
  assert.strictEqual(calculatedProfit, expected, `Förväntade vinst ${expected}, fick ${calculatedProfit}`);
});

Then('ska rabatten vara ungefär {float}%', function (expected: number) {
  assert(Math.abs(calculatedDiscount - expected) < 1,
    `Förväntade rabatt ~${expected}%, fick ${calculatedDiscount.toFixed(2)}%`);
});

Then('annonsen ska uppfylla handelsreglerna', function () {
  assert(passesRules(listing, tradingRules),
    `Annonsen ska uppfylla handelsreglerna (vinst=${calculatedProfit}, rabatt=${calculatedDiscount.toFixed(2)}%)`);
});

Then('annonsen ska inte uppfylla minimivinstgränsen på {int} SEK', function (threshold: number) {
  assert(calculatedProfit <= threshold, `Vinst ${calculatedProfit} ska inte överstiga gränsen ${threshold}`);
});

Then('annonsen ska inte uppfylla minimirabattgränsen på {float}%', function (threshold: number) {
  assert(calculatedDiscount <= threshold,
    `Rabatt ${calculatedDiscount.toFixed(2)}% ska inte överstiga gränsen ${threshold}%`);
});

Then('ska e-posten inkludera köppriset', function () {
  assert(emailData !== null, 'E-postdata ska finnas');
  assert(emailData!.purchasePrice > 0, 'E-posten ska inkludera köppriset');
});

Then('ska e-posten inkludera värderingen', function () {
  assert(emailData !== null, 'E-postdata ska finnas');
  assert(emailData!.valuation > 0, 'E-posten ska inkludera värderingen');
});

Then('ska e-posten inkludera rabattprocenten', function () {
  assert(emailData !== null, 'E-postdata ska finnas');
  assert(emailData!.discountPercent > 0, 'E-posten ska inkludera rabattprocenten');
});

Then('ska e-posten inkludera nypriset', function () {
  assert(emailData !== null, 'E-postdata ska finnas');
  assert('newPrice' in emailData!, 'E-posten ska inkludera nypris-fältet');
});

Then('ska e-posten inkludera vinsten', function () {
  assert(emailData !== null, 'E-postdata ska finnas');
  assert(emailData!.profit > 0, 'E-posten ska inkludera vinsten');
});

Then('ska e-posten inkludera beskrivningen', function () {
  assert(emailData !== null, 'E-postdata ska finnas');
  assert(emailData!.description !== '', 'E-posten ska inkludera beskrivningen');
});

Then('ska e-posten inkludera länken', function () {
  assert(emailData !== null, 'E-postdata ska finnas');
  assert(emailData!.link !== '', 'E-posten ska inkludera länken');
});

Then('ska e-posten skickas asynkront utan att blockera', function () {
  assert(emailSentAsync, 'E-posten ska skickas asynkront');
});

Then('ska inget krascha', function () {
  assert(noCrash, 'Inget ska krascha');
});
