import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import EmptyPlaceholder from './EmptyPlaceholder.vue';

const global = { mocks: { $t: (key: string) => key } };

describe('EmptyPlaceholder', () => {
  it('renders the default icon when no icon prop is given', () => {
    const wrapper = mount(EmptyPlaceholder, { global });
    expect(wrapper.find('i').classes()).toContain('pi-plus');
  });

  it('renders the provided icon class', () => {
    const wrapper = mount(EmptyPlaceholder, {
      global,
      props: { icon: 'pi-inbox' },
    });
    expect(wrapper.find('i').classes()).toContain('pi-inbox');
  });

  it('renders the i18n key as label when no label prop is given', () => {
    const wrapper = mount(EmptyPlaceholder, { global });
    // $t is mocked to return the key itself
    expect(wrapper.text()).toContain('globals.messages.emptyState');
  });

  it('renders the provided label text', () => {
    const wrapper = mount(EmptyPlaceholder, {
      global,
      props: { label: 'No campaigns yet' },
    });
    expect(wrapper.text()).toContain('No campaigns yet');
  });

  it('renders inside a <section> element', () => {
    const wrapper = mount(EmptyPlaceholder, { global });
    expect(wrapper.element.tagName).toBe('SECTION');
  });
});
