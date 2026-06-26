import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import CopyText from './CopyText.vue';
import { testI18n } from '../test/i18n';

const mockToast = vi.fn();

// Mock the composable so useGlobal() returns a controlled $utils stub.
vi.mock('../composables/useGlobal', () => ({
  useGlobal: () => ({ $utils: { toast: mockToast } }),
}));

// useI18n() needs the real i18n plugin installed.
const global = { plugins: [testI18n] };

beforeEach(() => {
  mockToast.mockClear();
  // happy-dom doesn't implement document.execCommand — define it before spying.
  Object.defineProperty(document, 'execCommand', {
    value: vi.fn(() => true),
    writable: true,
    configurable: true,
  });
});

describe('CopyText', () => {
  it('renders the text prop as visible content by default', () => {
    const wrapper = mount(CopyText, { global, props: { text: 'https://example.com' } });
    expect(wrapper.text()).toContain('https://example.com');
  });

  it('hides the text when hideText is true', () => {
    const wrapper = mount(CopyText, {
      global,
      props: { text: 'secret', hideText: true },
    });
    expect(wrapper.text()).not.toContain('secret');
  });

  it('always renders the copy icon', () => {
    const wrapper = mount(CopyText, { global, props: { text: 'hi' } });
    expect(wrapper.find('i.pi-copy').exists()).toBe(true);
  });

  it('calls document.execCommand("copy") when clicked', async () => {
    const wrapper = mount(CopyText, { global, props: { text: 'hello' } });
    await wrapper.find('a').trigger('click');
    expect(document.execCommand).toHaveBeenCalledWith('copy');
  });

  it('shows a toast after copying', async () => {
    const wrapper = mount(CopyText, { global, props: { text: 'hello' } });
    await wrapper.find('a').trigger('click');
    expect(mockToast).toHaveBeenCalledOnce();
  });
});
