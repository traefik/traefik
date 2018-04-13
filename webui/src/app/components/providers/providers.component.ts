import { Component, OnInit, OnDestroy } from '@angular/core';
import { ApiService } from '../../services/api.service';
import { Subscription } from 'rxjs/Subscription';

@Component({
  selector: 'app-providers',
  templateUrl: 'providers.component.html'
})
export class ProvidersComponent implements OnInit, OnDestroy {
  sub: Subscription;
  keys: string[];
  data: any;
  providers: any;
  tab: string;
  keyword: string;

  constructor(private apiService: ApiService) { }

  ngOnInit() {
    this.keyword = '';
    this.sub = this.apiService.fetchProviders().subscribe(data => {
      this.data = data;
      this.providers = data;
      this.keys = Object.keys(this.providers);
      this.tab = this.keys[0];
    });
  }

  filter(): void {
    const keyword = this.keyword.toLowerCase();
    this.providers = Object.keys(this.data)
      .filter(value => value !== 'acme' && value !== 'ACME')
      .reduce((acc, curr) => {
        return Object.assign(acc, {
          [curr]: {
            backends: this.data[curr].backends.filter(d => d.id.toLowerCase().includes(keyword)),
            frontends: this.data[curr].frontends.filter(d => d.id.toLowerCase().includes(keyword))
          }
        });
      }, {});
  }

  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }
}
