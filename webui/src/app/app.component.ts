import { Component, Inject } from '@angular/core';
import { DOCUMENT } from '@angular/common';

@Component({
  selector: 'app-root',
  template: `
    <main class="wip">
      <img src="./assets/images/traefik.logo.svg" alt="logo" />
      <header>
        <h1 class="title">
          <i class="fa fa-exclamation-triangle"></i>
          Work in progress...
        </h1>
        <p>
          In the meantime, you can review your current configuration by using
          the
          <a href="{{ href }}/api/rawdata">{{ href }}/api/rawdata</a> endpoint
          <br /><br />
          Also, please keep your <i class="fa fa-eye"></i> on our
          <a href="https://docs.traefik.io/v2.0/operations/dashboard/"
            >documentation</a
          >
          to stay informed
        </p>
        <p></p>
      </header>
    </main>
  `
})
export class AppComponent {
  public href: string;

  constructor(@Inject(DOCUMENT) private document: Document) {
    this.href = this.document.location.origin;
  }
}
