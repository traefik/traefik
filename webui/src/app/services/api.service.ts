import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/add/observable/empty';
import 'rxjs/add/observable/of';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/retry';

export interface ProviderType {
  [provider: string]: {
    backends: any;
    frontends: any;
  };
}

@Injectable()
export class ApiService {
  headers: HttpHeaders;

  constructor(private http: HttpClient) {
    this.headers = new HttpHeaders({
      'Access-Control-Allow-Origin': '*'
    });
  }

  fetchVersion(): Observable<any> {
    return this.http.get(`./api/version`, { headers: this.headers })
      .retry(4)
      .catch((err: HttpErrorResponse) => {
        console.error(`[version] returned code ${err.status}, body was: ${err.error}`);
        return Observable.empty<any>();
      });
  }

  fetchHealthStatus(): Observable<any> {
    return this.http.get(`./health`, { headers: this.headers })
      .retry(2)
      .catch((err: HttpErrorResponse) => {
        console.error(`[health] returned code ${err.status}, body was: ${err.error}`);
        return Observable.empty<any>();
      });
  }

  fetchProviders(): Observable<any> {
    return this.http.get(`./api/providers`, { headers: this.headers })
      .retry(2)
      .catch((err: HttpErrorResponse) => {
        console.error(`[providers] returned code ${err.status}, body was: ${err.error}`);
        return Observable.of<any>({});
      })
      .map(this.parseProviders);
  }

  parseProviders(data: any): ProviderType {
    return Object.keys(data)
      .filter(value => value !== 'acme' && value !== 'ACME')
      .reduce((acc, curr) => {
      acc[curr] = {
        backends: Object.keys(data[curr].backends || {}).map(key => {
          data[curr].backends[key].id = key;
          data[curr].backends[key].servers = Object.keys(data[curr].backends[key].servers || {}).map(server => {
            return {
              title: server,
              url: data[curr].backends[key].servers[server].url,
              weight: data[curr].backends[key].servers[server].weight
            };
          });

          return data[curr].backends[key];
        }),
        frontends: Object.keys(data[curr].frontends || {}).map(key => {
          data[curr].frontends[key].id = key;
          data[curr].frontends[key].routes = Object.keys(data[curr].frontends[key].routes || {}).map(route => {
            return {
              title: route,
              rule: data[curr].frontends[key].routes[route].rule
            };
          });

          return data[curr].frontends[key];
        }),
      };

      return acc;
    }, {});
  }
}
