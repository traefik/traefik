(function () {
  var BANNER_ID = 'migration-doc-banner';
  var SESSION_KEY = 'migration-doc-banner-dismissed';

  function createBanner() {
    if (document.getElementById(BANNER_ID)) return;
    if (sessionStorage.getItem(SESSION_KEY)) return;

    var banner = document.createElement('div');
    banner.id = BANNER_ID;
    banner.innerHTML =
      '<p><strong>Moving from ingress-nginx?</strong></p>' +
      '<p>No need to start over. Traefik supports your existing ingress-nginx annotations as-is &mdash; no rewrites, no downtime.</p>' +
      '<p>See our <a href="/traefik/migrate/nginx-to-traefik/">migration guide</a> and <a href="/traefik/reference/routing-configuration/kubernetes/ingress-nginx/">annotation reference</a> to get started.</p>' +
      '<button id="migration-doc-banner-close" aria-label="Dismiss banner">&times;</button>';

    var target =
      document.querySelector('.md-content__inner') ||
      document.querySelector('.md-main__inner') ||
      document.querySelector('article') ||
      document.querySelector('main');

    if (target) {
      target.insertBefore(banner, target.firstChild);
    }

    document.getElementById('migration-doc-banner-close').addEventListener('click', function () {
      banner.remove();
      sessionStorage.setItem(SESSION_KEY, '1');
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', createBanner);
  } else {
    createBanner();
  }
})();
