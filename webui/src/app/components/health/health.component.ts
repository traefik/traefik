import {
  ChangeDetectionStrategy,
  Component,
  OnDestroy,
  OnInit
} from '@angular/core';
import { distanceInWordsStrict, format, subSeconds } from 'date-fns';
import * as _ from 'lodash';
import {
  delay,
  distinctUntilChanged,
  filter,
  map,
  repeatWhen,
  retry,
  pluck,
  share
} from 'rxjs/operators';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-health',
  templateUrl: 'health.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class HealthComponent implements OnInit, OnDestroy {
  constructor(private readonly apiService: ApiService) {}

  private readonly source$ = this.apiService.fetchHealthStatus().pipe(
    repeatWhen(obs$ => obs$.pipe(delay(3000))),
    retry(),
    filter(data => Boolean(data)),
    share()
  );

  readonly data$ = this.source$.pipe(
    map(data => {
      return {
        pid: data.pid,

        uptime: distanceInWordsStrict(
          subSeconds(new Date(), data.uptime_sec),
          new Date()
        ),

        uptimeSince: format(
          subSeconds(new Date(), data.uptime_sec),
          'YYYY-MM-DD HH:mm:ss Z'
        ),

        totalResponseTime: distanceInWordsStrict(
          subSeconds(new Date(), data.total_response_time_sec),
          new Date()
        ),

        exactTotalResponseTime: data.total_response_time,
        averageResponseTime:
          Math.floor(data.average_response_time_sec * 1000) + ' ms',
        exactAverageResponseTime: data.average_response_time,
        codeCount: data.count,
        totalCodeCount: data.total_count
      };
    })
  );

  readonly chartValue$ = this.source$.pipe(
    map(({ average_response_time_sec: count, time: date }) => ({ count, date }))
  );

  readonly statusCodeValue$ = this.source$.pipe(
    map(({ total_status_code_count }) => _.toPairs(total_status_code_count)),
    map(pairs => pairs.map(([code, count]) => ({ code, count })))
  );

  readonly recentErrors$ = this.source$.pipe(
    pluck('recent_errors'),
    filter(errors => errors instanceof Array),
    map(_.cloneDeep),
    distinctUntilChanged((foo, bar) => _.isEqual(foo, bar))
  );

  ngOnInit() {}

  ngOnDestroy() {}

  trackRecentErrors(_index: number, item: RawHealth['recent_errors'][0]) {
    return item.status_code + item.method + item.host + item.path + item.time;
  }
}

export interface RawHealth {
  pid: number;
  time: string;

  count: number;
  total_count: number;

  average_response_time: string;
  average_response_time_sec: number;

  total_response_time: string;
  total_response_time_sec: number;

  uptime: string;
  uptime_sec: number;
  unixtime: number;

  status_code_count: Record<string, number>;
  total_status_code_count: Record<string, number>;

  recent_errors?: Array<{
    status: string;
    status_code: number;
    method: string;
    host: string;
    path: string;
    time: number;
  }>;
}
