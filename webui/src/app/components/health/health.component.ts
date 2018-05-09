import { Component, OnInit, OnDestroy } from '@angular/core';
import { ApiService } from '../../services/api.service';
import { Observable } from 'rxjs/Observable';
import { Subscription } from 'rxjs/Subscription';
import 'rxjs/add/observable/timer';
import 'rxjs/add/operator/timeInterval';
import 'rxjs/add/operator/mergeMap';
import 'rxjs/add/operator/map';
import { format, distanceInWordsStrict, subSeconds } from 'date-fns';

@Component({
  selector: 'app-health',
  templateUrl: 'health.component.html'
})
export class HealthComponent implements OnInit, OnDestroy {
  sub: Subscription;
  recentErrors: any;
  pid: number;
  uptime: string;
  uptimeSince: string;
  averageResponseTime: string;
  totalResponseTime: string;
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
          this.recentErrors = data.recent_errors;
          this.chartValue = { count: data.average_response_time_sec, date: data.time };
          this.statusCodeValue = Object.keys(data.total_status_code_count)
            .map(key => ({ code: key, count: data.total_status_code_count[key] }));

          this.pid = data.pid;
          this.uptime = distanceInWordsStrict(subSeconds(new Date(), data.uptime_sec), new Date());
          this.uptimeSince = format(subSeconds(new Date(), data.uptime_sec), 'MM/DD/YYYY HH:mm:ss');
          const re = /(.*\.\d{3})\d+(.*)/    // To display up to 3 digits of precision in durations
          this.totalResponseTime = data.total_response_time.replace(re, '$1$2');
          this.averageResponseTime = data.average_response_time.replace(re, '$1$2');
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
}
