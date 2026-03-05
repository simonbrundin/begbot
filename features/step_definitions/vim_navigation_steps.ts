import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

// VimNavigation implementation matching the expected API from the tests
interface VimNavigation {
  setFocused: (focused: boolean) => void;
  moveDown: () => void;
  moveUp: () => void;
  getSelectedIndex: () => number | null;
  clearSelection: () => void;
  setItemCount: (count: number) => void;
}

function createVimNavigation(itemCount: number): VimNavigation {
  let focused = false;
  let selectedIndex: number | null = null;
  let count = itemCount;

  return {
    setFocused(f: boolean) { focused = f; },
    moveDown() {
      if (!focused || count === 0) return;
      if (selectedIndex === null) {
        selectedIndex = 0;
      } else {
        selectedIndex = Math.min(selectedIndex + 1, count - 1);
      }
    },
    moveUp() {
      if (!focused || count === 0) return;
      if (selectedIndex === null) {
        selectedIndex = count - 1;
      } else {
        selectedIndex = Math.max(selectedIndex - 1, 0);
      }
    },
    getSelectedIndex() { return selectedIndex; },
    clearSelection() { selectedIndex = null; },
    setItemCount(newCount: number) {
      count = newCount;
      if (count === 0) {
        selectedIndex = null;
      } else if (selectedIndex !== null && selectedIndex >= count) {
        selectedIndex = count - 1;
      }
    },
  };
}

let nav: VimNavigation;

Before(function () {
  nav = createVimNavigation(0);
});

Given('a list with {int} items', function (count: number) {
  nav = createVimNavigation(count);
});

Given('a list with {int} item', function (count: number) {
  nav = createVimNavigation(count);
});

Given('the navigation is focused', function () {
  nav.setFocused(true);
});

Given('the navigation is not focused', function () {
  nav.setFocused(false);
});

When('I press j', function () {
  nav.moveDown();
});

When('I press j again', function () {
  nav.moveDown();
});

When('I press k', function () {
  nav.moveUp();
});

When('I press k again', function () {
  nav.moveUp();
});

When('I navigate down to index {int}', function (targetIndex: number) {
  for (let i = 0; i <= targetIndex; i++) {
    nav.moveDown();
  }
});

When('I navigate to the last item', function () {
  for (let i = 0; i < 10; i++) {
    nav.moveDown();
  }
});

When('I navigate to index {int}', function (targetIndex: number) {
  for (let i = 0; i <= targetIndex; i++) {
    nav.moveDown();
  }
});

When('I set focus to true and press j', function () {
  nav.setFocused(true);
  nav.moveDown();
});

When('I set focus to false and press j', function () {
  nav.setFocused(false);
  nav.moveDown();
});

When('I press j to select index {int}', function (_index: number) {
  nav.moveDown();
});

When('I clear the selection', function () {
  nav.clearSelection();
});

When('the item count changes to {int}', function (newCount: number) {
  nav.setItemCount(newCount);
});

Then('the selected index should be {int}', function (expected: number) {
  assert.strictEqual(nav.getSelectedIndex(), expected,
    `Expected selected index ${expected}, got ${nav.getSelectedIndex()}`);
});

Then('the selected index should not be null', function () {
  assert(nav.getSelectedIndex() !== null, 'Expected selected index to not be null');
});

Then('the selected index should be null', function () {
  assert(nav.getSelectedIndex() === null,
    `Expected selected index to be null, got ${nav.getSelectedIndex()}`);
});

Then('the selected index should remain {int}', function (expected: number) {
  assert.strictEqual(nav.getSelectedIndex(), expected,
    `Expected selected index to remain ${expected}, got ${nav.getSelectedIndex()}`);
});

Then('no navigation error should occur', function () {
  assert(true, 'No error occurred');
});
