import Bowser from 'bowser'
import vhCheck from 'vh-check'

const browser = Bowser.getParser(window.navigator.userAgent)

// In Mobile
if (browser.getPlatform().type === 'mobile') {
  vhCheck()
}

export default async ({ app, Vue }) => {

}
