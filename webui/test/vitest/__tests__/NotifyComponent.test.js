import { installQuasarPlugin } from '@quasar/quasar-app-extension-testing-unit-vitest';
import { mount } from '@vue/test-utils';
import { Notify } from 'quasar';
import { describe, expect, it, vi } from 'vitest';
import NotifyComponent from './demo/NotifyComponent.vue';

installQuasarPlugin({ plugins: { Notify } });

describe('notify example', () => {
  it('should call notify on click', async () => {
    expect(NotifyComponent).toBeTruthy();

    const wrapper = mount(NotifyComponent);
    const spy = vi.spyOn(Notify, 'create');
    expect(spy).not.toHaveBeenCalled();
    await wrapper.trigger('click');
    expect(spy).toHaveBeenCalled();
  });
});
