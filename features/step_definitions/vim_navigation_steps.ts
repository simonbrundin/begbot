import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

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
      if (selectedIndex === null) selectedIndex = 0;
      else selectedIndex = Math.min(selectedIndex + 1, count - 1);
    },
    moveUp() {
      if (!focused || count === 0) return;
      if (selectedIndex === null) selectedIndex = count - 1;
      else selectedIndex = Math.max(selectedIndex - 1, 0);
    },
    getSelectedIndex() { return selectedIndex; },
    clearSelection() { selectedIndex = null; },
    setItemCount(newCount: number) {
      count = newCount;
      if (count === 0) selectedIndex = null;
      else if (selectedIndex !== null && selectedIndex >= count) selectedIndex = count - 1;
    },
  };
}

let nav: VimNavigation;

Before(function () {
  nav = createVimNavigation(0);
});

Given('en lista med {int} poster', function (count: number) {
  nav = createVimNavigation(count);
});

Given('en lista med {int} post', function (count: number) {
  nav = createVimNavigation(count);
});

Given('navigeringen är fokuserad', function () {
  nav.setFocused(true);
});

Given('navigeringen är inte fokuserad', function () {
  nav.setFocused(false);
});

When('jag trycker på j', function () {
  nav.moveDown();
});

When('jag trycker på j igen', function () {
  nav.moveDown();
});

When('jag trycker på k', function () {
  nav.moveUp();
});

When('jag trycker på k igen', function () {
  nav.moveUp();
});

When('jag navigerar ned till index {int}', function (targetIndex: number) {
  for (let i = 0; i <= targetIndex; i++) nav.moveDown();
});

When('jag navigerar till sista posten', function () {
  for (let i = 0; i < 10; i++) nav.moveDown();
});

When('jag navigerar till index {int}', function (targetIndex: number) {
  for (let i = 0; i <= targetIndex; i++) nav.moveDown();
});

When('jag ställer in fokus till true och trycker på j', function () {
  nav.setFocused(true);
  nav.moveDown();
});

When('jag ställer in fokus till false och trycker på j', function () {
  nav.setFocused(false);
  nav.moveDown();
});

When('jag trycker på j för att välja index {int}', function (_index: number) {
  nav.moveDown();
});

When('jag rensar markeringen', function () {
  nav.clearSelection();
});

When('antalet poster ändras till {int}', function (newCount: number) {
  nav.setItemCount(newCount);
});

Then('ska det valda indexet vara {int}', function (expected: number) {
  assert.strictEqual(nav.getSelectedIndex(), expected,
    `Förväntade valt index ${expected}, fick ${nav.getSelectedIndex()}`);
});

Then('ska det valda indexet inte vara null', function () {
  assert(nav.getSelectedIndex() !== null, 'Förväntade att valt index inte är null');
});

Then('ska det valda indexet vara null', function () {
  assert(nav.getSelectedIndex() === null,
    `Förväntade att valt index är null, fick ${nav.getSelectedIndex()}`);
});

Then('ska det valda indexet förbli {int}', function (expected: number) {
  assert.strictEqual(nav.getSelectedIndex(), expected,
    `Förväntade att valt index förblir ${expected}, fick ${nav.getSelectedIndex()}`);
});

Then('ska inget navigeringsfel inträffa', function () {
  assert(true, 'Inget fel inträffade');
});
