import { Component, OnDestroy, OnInit } from '@angular/core';
import { distanceInWordsStrict, format, subSeconds } from 'date-fns';
import * as _ from 'lodash';
import 'rxjs/add/observable/timer';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/mergeMap';
import 'rxjs/add/operator/timeInterval';
import { Observable } from 'rxjs/Observable';
import { Subscription } from 'rxjs/Subscription';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-health',
  templateUrl: 'health.component.html'
})
export class HealthComponent implements OnInit, OnDestroy {
  sub: Subscription;
  recentErrors: any;
  previousRecentErrors: any;
  pid: number;
  uptime: string;
  uptimeSince: string;
  averageResponseTime: string;
  exactAverageResponseTime: string;
  totalResponseTime: string;
  exactTotalResponseTime: string;
  codeCount: number;
  totalCodeCount: number;
  chartValue: any;
  statusCodeValue: any;

  constructor(private apiService: ApiService) { }

  ngOnInit() {
    this.sub = Observable.timer(0, 3000)
      .timeInterval()
      .mergeMap(() => this.apiService.fetchHealthStatus())
      .subscribe(data => {
        if (data) {
          if (!_.isEqual(this.previousRecentErrors, data.recent_errors)) {
            this.previousRecentErrors = _.cloneDeep(data.recent_errors);
            this.recentErrors = data.recent_errors;
          }

          this.chartValue = {count: data.average_response_time_sec, date: data.time};
          this.statusCodeValue = Object.keys(data.total_status_code_count)
            .map(key => ({code: key, count: data.total_status_code_count[key]}));

          this.pid = data.pid;
          this.uptime = distanceInWordsStrict(subSeconds(new Date(), data.uptime_sec), new Date());
          this.uptimeSince = format(subSeconds(new Date(), data.uptime_sec), 'YYYY-MM-DD HH:mm:ss Z');
          this.totalResponseTime = distanceInWordsStrict(subSeconds(new Date(), data.total_response_time_sec), new Date());
          this.exactTotalResponseTime = data.total_response_time;
          this.averageResponseTime = Math.floor(data.average_response_time_sec * 1000) + ' ms';
          this.exactAverageResponseTime = data.average_response_time;
          this.codeCount = data.count;
          this.totalCodeCount = data.total_count;
        }
      });
  }

  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }

  trackRecentErrors(index, item): string {
    return item.status_code + item.method + item.host + item.path + item.time;
  }
}
