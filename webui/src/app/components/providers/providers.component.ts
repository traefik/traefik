import { Component, OnDestroy, OnInit } from '@angular/core';
import * as _ from 'lodash';
import { Observable } from 'rxjs/Observable';
import { Subscription } from 'rxjs/Subscription';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-providers',
  templateUrl: 'providers.component.html'
})
export class ProvidersComponent implements OnInit, OnDestroy {
  sub: Subscription;
  maxItem: number;
  keys: string[];
  previousKeys: string[];
  previousData: any;
  providers: any;
  tab: string;
  keyword: string;

  constructor(private apiService: ApiService) { }

  ngOnInit() {
    this.maxItem = 100;
    this.keyword = '';
    this.sub = Observable.timer(0, 2000)
      .timeInterval()
      .mergeMap(() => this.apiService.fetchProviders())
      .subscribe(data => {
        if (!_.isEqual(this.previousData, data)) {
          this.previousData = _.cloneDeep(data);
          this.providers = data;

          const keys = Object.keys(this.providers);
          if (!_.isEqual(this.previousKeys, keys)) {
            this.keys = keys;

            // keep current tab or set to the first tab
            if (!this.tab || (this.tab && !this.keys.includes(this.tab))) {
              this.tab = this.keys[0];
            }
          }
        }
      });
  }

  trackItem(tab): (index, item) => string {
    return (index, item): string => tab + '-' + item.id;
  }

  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }
}
