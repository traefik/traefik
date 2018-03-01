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
  pid: number;
  uptime: string;
  uptimeSince: string;
  averageResponseTime: string;
  totalResponseTime: string;
  codeCount: number;
  totalCodeCount: number;
  healthData: any[] = [];
  chartValue: any;

  constructor(private apiService: ApiService) { }

  ngOnInit() {
    this.sub = Observable.timer(0, 3000)
      .timeInterval()
      .mergeMap(() => this.apiService.fetchHealthStatus())
      .subscribe(data => {
        this.chartValue = { count: data.average_response_time_sec, date: data.time };
        this.pid = data.pid;
        this.uptime = distanceInWordsStrict(subSeconds(new Date(), data.uptime_sec), new Date());
        this.uptimeSince = format(subSeconds(new Date(), data.uptime_sec), 'MM/DD/YYYY HH:mm:ss');
        this.totalResponseTime = data.total_response_time;
        this.averageResponseTime = data.average_response_time;
        this.codeCount = data.count;
        this.totalCodeCount = data.total_count;

        this.healthData.push({
          average_response_time: data.average_response_time_sec,
          count: data.count,
          status_code_count: data.status_code_count,
          total_status_code_count: data.total_status_code_count,
          time: data.time
        });
      });
  }

  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }
}
