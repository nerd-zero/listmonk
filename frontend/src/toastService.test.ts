import {
  describe, it, expect, vi, beforeEach,
} from 'vitest';
import {
  setToastInstance, setConfirmInstance, showToast, showConfirm,
} from './toastService';

// ---------------------------------------------------------------------------
// showToast
// ---------------------------------------------------------------------------

describe('showToast', () => {
  const mockAdd = vi.fn();

  beforeEach(() => {
    mockAdd.mockClear();
    setToastInstance({ add: mockAdd });
  });

  it('calls toast.add with the correct severity mapping for is-success', () => {
    showToast('Saved!', 'is-success', 3000);
    expect(mockAdd).toHaveBeenCalledWith({ severity: 'success', detail: 'Saved!', life: 3000 });
  });

  it('maps is-danger to error severity', () => {
    showToast('Something went wrong', 'is-danger', 4000);
    expect(mockAdd).toHaveBeenCalledWith({ severity: 'error', detail: 'Something went wrong', life: 4000 });
  });

  it('maps is-warning to warn severity', () => {
    showToast('Watch out', 'is-warning', 3000);
    expect(mockAdd).toHaveBeenCalledWith({ severity: 'warn', detail: 'Watch out', life: 3000 });
  });

  it('maps is-info to info severity', () => {
    showToast('FYI', 'is-info', 3000);
    expect(mockAdd).toHaveBeenCalledWith({ severity: 'info', detail: 'FYI', life: 3000 });
  });

  it('falls back to success severity for unknown type strings', () => {
    showToast('Hello', 'is-unknown', 3000);
    expect(mockAdd).toHaveBeenCalledWith({ severity: 'success', detail: 'Hello', life: 3000 });
  });

  it('passes the duration through as the life value', () => {
    showToast('Quick', 'is-success', 1500);
    expect(mockAdd).toHaveBeenCalledWith(expect.objectContaining({ life: 1500 }));
  });

  it('does not throw when no toast instance has been set', () => {
    setToastInstance(null as any);
    expect(() => showToast('Hello', 'is-success', 3000)).not.toThrow();
  });
});

// ---------------------------------------------------------------------------
// showConfirm
// ---------------------------------------------------------------------------

describe('showConfirm — with confirm instance', () => {
  const mockRequire = vi.fn();

  beforeEach(() => {
    mockRequire.mockClear();
    setConfirmInstance({ require: mockRequire });
  });

  it('calls confirmInstance.require with message and callbacks', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    showConfirm('Are you sure?', onConfirm, onCancel);

    expect(mockRequire).toHaveBeenCalledOnce();
    const opts = mockRequire.mock.calls[0][0];
    expect(opts.message).toBe('Are you sure?');
    expect(opts.accept).toBe(onConfirm);
    expect(opts.reject).toBe(onCancel);
  });

  it('sets a non-empty header', () => {
    showConfirm('Delete this?');
    const opts = mockRequire.mock.calls[0][0];
    expect(opts.header).toBeTruthy();
  });

  it('works when callbacks are omitted', () => {
    expect(() => showConfirm('Delete?')).not.toThrow();
  });
});

describe('showConfirm — without confirm instance (browser fallback)', () => {
  beforeEach(() => {
    setConfirmInstance(null as any);
    // happy-dom doesn't implement window.confirm — define it before spying.
    window.confirm = vi.fn();
  });

  it('calls the onConfirm callback when window.confirm returns true', () => {
    vi.mocked(window.confirm).mockReturnValueOnce(true);
    const onConfirm = vi.fn();
    showConfirm('Sure?', onConfirm);
    expect(onConfirm).toHaveBeenCalledOnce();
  });

  it('calls the onCancel callback when window.confirm returns false', () => {
    vi.mocked(window.confirm).mockReturnValueOnce(false);
    const onCancel = vi.fn();
    showConfirm('Sure?', undefined, onCancel);
    expect(onCancel).toHaveBeenCalledOnce();
  });
});
