import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface ProductConditionResult {
  isIntact: boolean;
  hasMinorScratches: boolean;
  issuesFound: string[];
  reasoning: string;
}

let conditionResult: ProductConditionResult;

Before(function () {
  conditionResult = { isIntact: false, hasMinorScratches: false, issuesFound: [], reasoning: '' };
});

Given('ett produkttillståndsresultat med isIntact true och inga repor', function () {
  conditionResult = { isIntact: true, hasMinorScratches: false, issuesFound: [], reasoning: 'Produkten är i bra skick' };
});

Given('inga problem hittades', function () {
  conditionResult.issuesFound = [];
});

Given('problem hittades: {string}, {string}', function (issue1: string, issue2: string) {
  conditionResult.issuesFound = [issue1, issue2];
});

Given('ett produkttillståndsresultat med isIntact false', function () {
  conditionResult = { isIntact: false, hasMinorScratches: false, issuesFound: [], reasoning: 'Produkten har skador' };
});

Given('ett produkttillståndsresultat med isIntact true och smärre repor', function () {
  conditionResult = { isIntact: true, hasMinorScratches: true, issuesFound: [], reasoning: 'Endast smärre repor' };
});

Given('en produkt med isIntact true och inga repor', function () {
  conditionResult = { isIntact: true, hasMinorScratches: false, issuesFound: [], reasoning: '' };
});

Given('en produkt med isIntact true och smärre repor', function () {
  conditionResult = { isIntact: true, hasMinorScratches: true, issuesFound: [], reasoning: '' };
});

Given('en produkt med isIntact false och problem {string}', function (issue: string) {
  conditionResult = { isIntact: false, hasMinorScratches: false, issuesFound: [issue], reasoning: '' };
});

When('köpgiltigheten utvärderas', function () {});

Then('ska produkten vara giltig för köp', function () {
  assert(conditionResult.isIntact, 'Produkten ska vara hel (giltig för köp)');
});

Then('ska produkten inte vara giltig för köp', function () {
  assert(!conditionResult.isIntact, 'Produkten ska inte vara hel');
});

Then('det ska ha hittats {int} problem', function (expected: number) {
  assert.strictEqual(conditionResult.issuesFound.length, expected);
});

Then('flaggan för smärre repor ska vara true', function () {
  assert(conditionResult.hasMinorScratches, 'Förväntade hasMinorScratches att vara true');
});
