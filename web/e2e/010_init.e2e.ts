import { browser, by, element } from 'protractor';

describe('Application Init', () => {

  it('should have correct title', () => {
    return browser.get('/')
      .then(() => browser.getTitle())
      .then(title => expect(title).toBe('Angular Webpack Seed'));
  });

});
