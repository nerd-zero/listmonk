import { describe, it, expect } from 'vitest';
import Utils from './utils';

// Instantiate with a null i18n — only tests for i18n-independent methods here.
const u = new Utils(null as never);

// ---------------------------------------------------------------------------
// camelString
// ---------------------------------------------------------------------------

describe('Utils.camelString', () => {
  it('converts snake_case to camelCase', () => {
    expect(u.camelString('hello_world')).toBe('helloWorld');
  });

  it('converts kebab-case to camelCase', () => {
    expect(u.camelString('hello-world')).toBe('helloWorld');
  });

  it('converts space-separated words to camelCase', () => {
    expect(u.camelString('hello world')).toBe('helloWorld');
  });

  it('handles multiple separators in sequence', () => {
    expect(u.camelString('foo_bar_baz')).toBe('fooBarBaz');
  });

  it('leaves already-camelCase strings unchanged', () => {
    expect(u.camelString('helloWorld')).toBe('helloWorld');
  });

  it('lowercases the first character', () => {
    expect(u.camelString('Hello')).toBe('hello');
  });
});

// ---------------------------------------------------------------------------
// camelKeys
// ---------------------------------------------------------------------------

describe('Utils.camelKeys', () => {
  it('converts top-level snake_case keys', () => {
    expect(u.camelKeys({ foo_bar: 1 })).toEqual({ fooBar: 1 });
  });

  it('recursively converts nested keys', () => {
    expect(u.camelKeys({ outer_key: { inner_key: 42 } })).toEqual({
      outerKey: { innerKey: 42 },
    });
  });

  it('handles arrays of objects', () => {
    expect(u.camelKeys([{ first_name: 'Alice' }, { first_name: 'Bob' }])).toEqual([
      { firstName: 'Alice' },
      { firstName: 'Bob' },
    ]);
  });

  it('leaves primitive values unchanged', () => {
    expect(u.camelKeys('hello')).toBe('hello');
    expect(u.camelKeys(42)).toBe(42);
    expect(u.camelKeys(null)).toBeNull();
  });

  it('does not convert keys that match the testFunc exclusion', () => {
    const skip = (keyPath: string) => !keyPath.endsWith('.keep_as_is');
    const result = u.camelKeys({ keep_as_is: 1, convert_me: 2 }, skip) as any;
    expect(result.keep_as_is).toBe(1);
    expect(result.convertMe).toBe(2);
  });

  it('handles empty objects and arrays', () => {
    expect(u.camelKeys({})).toEqual({});
    expect(u.camelKeys([])).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// niceNumber
// ---------------------------------------------------------------------------

describe('Utils.niceNumber', () => {
  it('returns the number as-is for values below 10 000', () => {
    expect(u.niceNumber(0)).toBe(0);
    expect(u.niceNumber(999)).toBe(999);
    expect(u.niceNumber(9999)).toBe(9999);
  });

  it('abbreviates whole thousands with k suffix (no decimals)', () => {
    expect(u.niceNumber(10000)).toBe('10k');
    expect(u.niceNumber(50000)).toBe('50k');
  });

  it('abbreviates fractional thousands to 2 decimal places', () => {
    expect(u.niceNumber(15500)).toBe('15.50k');
    expect(u.niceNumber(12345)).toBe('12.35k');
  });

  it('abbreviates whole millions with m suffix (no decimals)', () => {
    expect(u.niceNumber(1_000_000)).toBe('1m');
    expect(u.niceNumber(5_000_000)).toBe('5m');
  });

  it('abbreviates fractional millions to 2 decimal places', () => {
    expect(u.niceNumber(2_500_000)).toBe('2.50m');
  });

  it('abbreviates whole billions with b suffix (no decimals)', () => {
    expect(u.niceNumber(1_000_000_000)).toBe('1b');
  });

  it('returns 0 for null and undefined', () => {
    expect(u.niceNumber(null)).toBe(0);
    expect(u.niceNumber(undefined)).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// parseQueryIDs
// ---------------------------------------------------------------------------

describe('Utils.parseQueryIDs', () => {
  it('parses a single string id', () => {
    expect(u.parseQueryIDs('5')).toEqual([5]);
  });

  it('parses a single numeric id', () => {
    expect(u.parseQueryIDs(7)).toEqual([7]);
  });

  it('parses an array of string ids', () => {
    expect(u.parseQueryIDs(['1', '2', '3'])).toEqual([1, 2, 3]);
  });

  it('parses an array of numeric ids', () => {
    expect(u.parseQueryIDs([10, 20, 30])).toEqual([10, 20, 30]);
  });

  it('returns an empty array for null', () => {
    expect(u.parseQueryIDs(null)).toEqual([]);
  });

  it('returns an empty array for undefined', () => {
    expect(u.parseQueryIDs(undefined)).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// escapeHTML
// ---------------------------------------------------------------------------

describe('Utils.escapeHTML', () => {
  it('escapes angle brackets', () => {
    expect(u.escapeHTML('<b>hi</b>')).toBe('&lt;b&gt;hi&lt;&#x2F;b&gt;');
  });

  it('escapes ampersands', () => {
    expect(u.escapeHTML('a & b')).toBe('a &amp; b');
  });

  it('escapes double quotes', () => {
    expect(u.escapeHTML('"hello"')).toBe('&quot;hello&quot;');
  });

  it('escapes single quotes', () => {
    expect(u.escapeHTML("it's")).toBe('it&#39;s');
  });

  it('leaves plain text unchanged', () => {
    expect(u.escapeHTML('hello world')).toBe('hello world');
  });
});

// ---------------------------------------------------------------------------
// titleCase
// ---------------------------------------------------------------------------

describe('Utils.titleCase', () => {
  it('capitalises the first letter and lowercases the rest', () => {
    expect(u.titleCase('hello')).toBe('Hello');
    expect(u.titleCase('WORLD')).toBe('World');
    expect(u.titleCase('hELLO')).toBe('Hello');
  });
});

// ---------------------------------------------------------------------------
// validateEmail
// ---------------------------------------------------------------------------

describe('Utils.validateEmail', () => {
  it('accepts valid email addresses', () => {
    expect(u.validateEmail('user@example.com')).toBeTruthy();
    expect(u.validateEmail('a@b.io')).toBeTruthy();
  });

  it('rejects strings without @', () => {
    expect(u.validateEmail('notanemail')).toBeFalsy();
  });
});

// ---------------------------------------------------------------------------
// getPref / setPref (via localStorage)
// ---------------------------------------------------------------------------

describe('Utils.getPref / setPref', () => {
  it('returns null when nothing has been set', () => {
    localStorage.clear();
    expect(u.getPref('some.key')).toBeNull();
  });

  it('persists and retrieves a preference value', () => {
    u.setPref('ui.tab', 3);
    expect(u.getPref('ui.tab')).toBe(3);
  });

  it('returns null for an unknown key even when prefs exist', () => {
    u.setPref('ui.tab', 3);
    expect(u.getPref('ui.unknown')).toBeNull();
  });

  it('overwrites an existing preference', () => {
    u.setPref('ui.tab', 1);
    u.setPref('ui.tab', 5);
    expect(u.getPref('ui.tab')).toBe(5);
  });
});
