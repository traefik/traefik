<template>
  <q-header class="shadow-1">
    <section class="app-section bg-primary text-white">
      <div class="app-section-wrap app-boxed app-boxed-xl">
        <q-toolbar class="row no-wrap items-center">
          <div class="q-pr-md logo">
            <img
              alt="logo"
              src="~assets/logo.svg"
            >
            <q-btn
              v-if="version"
              type="a"
              href="https://github.com/traefik/traefik/"
              target="_blank"
              stretch
              flat
              no-caps
              :label="version"
              class="btn-menu version"
            />
          </div>
          <q-tabs
            align="left"
            inline-label
            indicator-color="transparent"
            active-color="white"
            stretch
          >
            <q-route-tab
              to="/"
              icon="eva-home-outline"
              no-caps
              label="Dashboard"
            />
            <q-route-tab
              to="/http"
              icon="eva-globe-outline"
              no-caps
              label="HTTP"
            />
            <q-route-tab
              to="/tcp"
              icon="eva-globe-2-outline"
              no-caps
              label="TCP"
            />
            <q-route-tab
              to="/udp"
              icon="eva-globe-2-outline"
              no-caps
              label="UDP"
            />
            <q-btn
              type="a"
              href="https://plugins.traefik.io"
              target="_blank"
              flat
              no-caps
              class="btn-menu"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                xmlnsXlink="http://www.w3.org/1999/xlink"
                width="24"
                height="24"
                viewBox="0 0 24 24"
              >
                <path
                  fill="currentColor"
                  d="M11.2126113,3.1448713 C11.6312565,2.9520971 12.1102888,2.95151646 12.5277726,3.14429065 C12.5277726,3.1448713 12.5283533,3.1448713 12.5289339,3.1448713 L16.245063,4.87112936 C16.4448049,4.96345194 16.6056436,5.10454872 16.7420952,5.26480678 C16.7618372,5.28164549 16.7763533,5.30371001 16.7943533,5.32345194 C16.8251275,5.3640971 16.8576436,5.40300033 16.8843533,5.44712936 C16.8872565,5.45177452 16.8919017,5.45467775 16.8936436,5.45932291 L16.8936436,5.45932291 L16.8971275,5.47383904 C17.0202243,5.68461323 17.0962888,5.92093581 17.0968694,6.16945194 L17.0968694,6.16945194 L17.0968694,10.2821616 C17.0967394,10.405418 17.0798401,10.5266364 17.0478731,10.6432353 L17.1729339,10.6926777 C17.1729339,10.6932584 17.1735146,10.6932584 17.1740952,10.6932584 L17.1740952,10.6932584 L20.8902243,12.4195165 C21.0899662,12.511839 21.2508049,12.6529358 21.3872565,12.8131939 C21.4069984,12.8300326 21.4215146,12.8520971 21.4395146,12.871839 C21.4702888,12.9124842 21.5028049,12.9513874 21.5295146,12.9955165 C21.5324178,13.0001616 21.537063,13.0030648 21.5388049,13.00771 C21.5411275,13.0123552 21.5405468,13.017581 21.5422888,13.0222261 C21.6653855,13.2330003 21.7414501,13.4693229 21.7420307,13.717839 L21.7420307,13.717839 L21.7420307,17.8305487 C21.7414501,18.3810003 21.4064178,18.8908068 20.8884823,19.1300326 L20.8884823,19.1300326 L17.1740952,20.85571 C16.9644823,20.9532584 16.739192,21.0020326 16.5139017,21.0020326 C16.2880307,21.0020326 16.0627404,20.9532584 15.8525468,20.8551294 L15.8525468,20.8551294 L12.1369984,19.1294519 C12.0404689,19.0843184 11.9504046,19.0298283 11.8677687,18.9674749 C11.7859554,19.0308103 11.6952726,19.0851777 11.5981597,19.1300326 L11.5981597,19.1300326 L7.88377264,20.85571 C7.67415974,20.9532584 7.44886941,21.0020326 7.22357909,21.0020326 C6.99770812,21.0020326 6.7724178,20.9532584 6.56222425,20.8551294 L6.56222425,20.8551294 L2.84667587,19.1294519 C2.32757909,18.8867423 1.99545006,18.3734519 2.00004712,17.821839 L2.00004712,17.821839 L2.00004712,13.7184197 C2.00004712,13.4699036 2.07615974,13.233581 2.19867587,13.0233874 L2.19867587,13.0233874 L2.20274038,13.00771 C2.20506296,13.0030648 2.20970812,12.999581 2.21261135,12.9943552 C2.23815974,12.9525487 2.26893393,12.9148068 2.29912748,12.8759036 C2.31828877,12.8544197 2.33396619,12.8300326 2.35545006,12.8120326 C2.49190167,12.6517745 2.65332103,12.5112584 2.85306296,12.4189358 L2.85306296,12.4189358 L6.56745006,10.6932584 C6.60898751,10.6741316 6.65111943,10.6568968 6.69372673,10.6415552 C6.66069389,10.5235765 6.64415782,10.399515 6.64520841,10.2734519 L6.64520841,10.2734519 L6.64520841,6.17003259 C6.64520841,5.92151646 6.72132103,5.68519388 6.84383716,5.47500033 L6.84383716,5.47500033 L6.84790167,5.45932291 L6.84790167,5.45932291 L6.85777264,5.44596807 C6.88332103,5.40416162 6.91409522,5.36641968 6.94428877,5.32751646 C6.96345006,5.30603259 6.97912748,5.28164549 7.00061135,5.26364549 C7.13706296,5.10338742 7.29848232,4.9628713 7.49822425,4.87054872 L7.49822425,4.87054872 Z M3.16137076,14.0935165 L3.16137076,17.8259036 C3.16022425,17.931581 3.22757909,18.0256455 3.33732103,18.0767423 L3.33732103,18.0767423 L6.64525651,19.6125487 L6.64525651,15.7117745 L3.16137076,14.0935165 Z M12.4516933,14.0935165 L12.4516933,17.8259036 C12.4505468,17.931581 12.5179017,18.0256455 12.6276436,18.0767423 L12.6276436,18.0767423 L15.9355791,19.6125487 L15.9355791,15.7117745 L12.4516933,14.0935165 Z M11.2904178,14.0935165 L7.80654683,15.7117745 L7.80654683,19.6113874 L11.1104178,18.0761616 C11.2219017,18.0250648 11.2904178,17.9310003 11.2904178,17.8299681 L11.2904178,17.8299681 L11.2904178,14.0935165 Z M20.5807404,14.0935165 L17.0968694,15.7117745 L17.0968694,19.6113874 L20.4007404,18.0761616 C20.5122243,18.0250648 20.5807404,17.9310003 20.5807404,17.8299681 L20.5807404,17.8299681 L20.5807404,14.0935165 Z M7.22590167,11.7105487 C7.16725651,11.7105487 7.10861135,11.7227423 7.05577264,11.74771 L7.05577264,11.74771 L3.96151458,13.1842261 L7.22590167,14.7008713 L10.4897081,13.1842261 L7.39545006,11.74771 C7.34261135,11.7227423 7.28396619,11.7105487 7.22590167,11.7105487 Z M16.5162243,11.7105487 C16.4575791,11.7105487 16.3989339,11.7227423 16.3460952,11.74771 L16.3460952,11.74771 L13.2518372,13.1842261 L16.5162243,14.7008713 L19.7800307,13.1842261 L16.6857726,11.74771 C16.6329339,11.7227423 16.5742888,11.7105487 16.5162243,11.7105487 Z M7.80653206,6.54512936 L7.80653206,10.2775165 C7.80538554,10.3831939 7.87274038,10.4772584 7.98248232,10.5283552 L7.98248232,10.5283552 L11.2904178,12.0641616 L11.2904178,8.16338742 L7.80653206,6.54512936 Z M15.9355791,6.54512936 L12.4517081,8.16338742 L12.4517081,12.0630003 L15.7555791,10.5277745 C15.867063,10.4766777 15.9355791,10.3826132 15.9355791,10.281581 L15.9355791,10.281581 L15.9355791,6.54512936 Z M11.871063,4.16216162 C11.8124178,4.16216162 11.7537726,4.17435517 11.7009339,4.19932291 L11.7009339,4.19932291 L8.60667587,5.63583904 L11.871063,7.1524842 L15.1348694,5.63583904 L12.0406113,4.19932291 C11.9877726,4.17435517 11.9291275,4.16216162 11.871063,4.16216162 Z"
                />
              </svg>
              Plugins
            </q-btn>
          </q-tabs>
          <div class="right-menu">
            <q-tabs class="allow-overflow">
              <div
                v-if="!coreVersion.disableDashboardAd && hasHubButtonComponent"
                style="margin-right: 5px;"
              >
                <hub-button-app
                  v-if="$q.dark.isActive"
                  theme="dark"
                />
                <hub-button-app v-if="!$q.dark.isActive" />
              </div>
              <q-btn
                stretch
                flat
                no-caps
                icon="invert_colors"
                :label="`${$q.dark.isActive ? 'Light' : 'Dark'} theme`"
                class="btn-menu"
                @click="$q.dark.toggle()"
              />
              <q-btn
                stretch
                flat
                icon="eva-question-mark-circle-outline"
              >
                <q-menu
                  anchor="bottom left"
                  auto-close
                >
                  <q-item>
                    <q-btn
                      type="a"
                      :href="`https://doc.traefik.io/traefik/${parsedVersion}`"
                      target="_blank"
                      flat
                      color="accent"
                      align="left"
                      icon="eva-book-open-outline"
                      no-caps
                      label="Documentation"
                      class="btn-submenu full-width"
                    />
                  </q-item>
                  <q-separator />
                  <q-item>
                    <q-btn
                      type="a"
                      href="https://github.com/traefik/traefik/"
                      target="_blank"
                      flat
                      color="accent"
                      align="left"
                      icon="eva-github-outline"
                      no-caps
                      label="GitHub repository"
                      class="btn-submenu full-width"
                    />
                  </q-item>
                </q-menu>
              </q-btn>
            </q-tabs>
          </div>
        </q-toolbar>
      </div>
    </section>

    <section
      class="app-section text-black sub-nav"
      :class="{ 'bg-white': !$q.dark.isActive }"
    >
      <div class="app-section-wrap app-boxed app-boxed-xl">
        <slot />
      </div>
    </section>
  </q-header>
</template>

<script>
import { defineComponent } from 'vue'
import config from '../../../package'
import { mapActions, mapGetters } from 'vuex'

export default defineComponent({
  name: 'NavBar',
  data () {
    return {
      hasHubButtonComponent: false
    }
  },
  computed: {
    ...mapGetters('core', { coreVersion: 'version' }),
    version () {
      if (!this.coreVersion.Version) return null
      return /^(v?\d+\.\d+)/.test(this.coreVersion.Version)
        ? this.coreVersion.Version
        : this.coreVersion.Version.substring(0, 7)
    },
    parsedVersion () {
      if (!this.version) {
        return 'master'
      }
      if (this.version === 'dev') {
        return 'master'
      }
      const matches = this.version.match(/^(v?\d+\.\d+)/)
      return matches ? 'v' + matches[1] : 'master'
    },
    name () {
      return config.productName
    },
    disableDashboardAd () {
      return this.coreVersion.disableDashboardAd
    }
  },
  watch: {
    disableDashboardAd (newValue) {
      if (!newValue && customElements.get('hub-button-app') === undefined) {
        const hubButtonScript = document.createElement('script')
        hubButtonScript.async = true
        hubButtonScript.onerror = () => {
          const hubButtonScriptLocal = document.createElement('script')
          hubButtonScriptLocal.async = true
          hubButtonScriptLocal.onload = () => {
            this.hasHubButtonComponent = customElements.get('hub-button-app') !== undefined
          }
          // Sources: https://github.com/traefik/traefiklabs-hub-button-app
          hubButtonScriptLocal.src = 'traefiklabs-hub-button-app/main-v1.js'
          document.head.appendChild(hubButtonScriptLocal)
        }
        hubButtonScript.onload = () => {
          this.hasHubButtonComponent = customElements.get('hub-button-app') !== undefined
        }
        // Sources: https://github.com/traefik/traefiklabs-hub-button-app
        hubButtonScript.src = 'https://traefik.github.io/traefiklabs-hub-button-app/main-v1.js'
        document.head.appendChild(hubButtonScript)
      }
    }
  },
  created () {
    this.getVersion()
  },
  methods: {
    ...mapActions('core', { getVersion: 'getVersion' })
  }
})
</script>

<style scoped lang="scss">
  @import "../../css/sass/variables";

  .q-toolbar {
    min-height: 64px;
  }

  .body--dark {
    .sub-nav {
      background-color: #0e204c;
    }
  }

  .q-item--dark {
    background: var(--q-color-dark);
  }

  .logo {
    display: flex;
    align-items: center;

    img {
      height: 24px;
      margin-right: 10px;
    }

    .version {
      min-height: inherit;
      line-height:  inherit;
      padding: 0 4px;
    }
  }

  .q-tabs {
    color: rgba( $app-text-white, .4 );
    :deep(.q-tabs__content) {
      .q-tab__content{
        min-width: 100%;
        .q-tab__label {
          font-size: 16px;
          font-weight: 600;
        }
      }
    }
  }

  .right-menu {
    flex: 1;
    height: 64px;
    display: flex;
    justify-content: flex-end;
  }

  .btn-menu {
    color: rgba( $app-text-white, .4 );
    font-size: 16px;
    font-weight: 600;
  }

  .q-item {
    padding: 0;
  }

  .btn-submenu {
    font-weight: 700;
    align-items: flex-start;
  }

  .allow-overflow {
    :deep(.q-tabs__content) {
      overflow: visible !important;
    }
  }

</style>
