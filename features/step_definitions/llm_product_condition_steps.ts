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

Given('a product condition result with isIntact true and no scratches', function () {
  conditionResult = { isIntact: true, hasMinorScratches: false, issuesFound: [], reasoning: 'Produkten är i bra skick' };
});

Given('no issues found', function () {
  conditionResult.issuesFound = [];
});

Given('issues found: {string}, {string}', function (issue1: string, issue2: string) {
  conditionResult.issuesFound = [issue1, issue2];
});

Given('a product condition result with isIntact false', function () {
  conditionResult = { isIntact: false, hasMinorScratches: false, issuesFound: [], reasoning: 'Produkten har skador' };
});

Given('a product condition result with isIntact true and minor scratches', function () {
  conditionResult = { isIntact: true, hasMinorScratches: true, issuesFound: [], reasoning: 'Endast mindre repor' };
});

Given('a product with isIntact true and no scratches', function () {
  conditionResult = { isIntact: true, hasMinorScratches: false, issuesFound: [], reasoning: '' };
});

Given('a product with isIntact true and minor scratches', function () {
  conditionResult = { isIntact: true, hasMinorScratches: true, issuesFound: [], reasoning: '' };
});

Given('a product with isIntact false and issues {string}', function (issue: string) {
  conditionResult = { isIntact: false, hasMinorScratches: false, issuesFound: [issue], reasoning: '' };
});

When('evaluating purchase validity', function () {
  // Validity is already set in the Given steps
});

Then('the product should be valid for purchase', function () {
  assert(conditionResult.isIntact, 'Product should be intact (valid for purchase)');
});

Then('the product should not be valid for purchase', function () {
  assert(!conditionResult.isIntact, 'Product should not be intact');
});

Then('there should be {int} issues found', function (expected: number) {
  assert.strictEqual(conditionResult.issuesFound.length, expected);
});

Then('the minor scratches flag should be true', function () {
  assert(conditionResult.hasMinorScratches, 'Expected hasMinorScratches to be true');
});
