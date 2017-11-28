import { Injectable } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/add/observable/of';
import 'rxjs/add/operator/map';
import { fakeData } from './fake-data';

export interface ProviderType {
  [provider: string]: {
    backends: any;
    frontends: any;
  };
}

@Injectable()
export class ApiService {
  headers: Headers;

  constructor(private http: Http) {
    this.headers = new Headers({
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'GET',
      'Access-Control-Allow-Origin': '*'
    });
  }

  fetchVersion(): Observable<any> {
    return this.http.get(`/api/version`, { headers: this.headers })
      .map(res => res.json());
  }

  fetchHealthStatus(): Observable<any> {
    return this.http.get(`/health`, { headers: this.headers })
      .map(res => res.json());
  }

  fetchProviders(): Observable<any> {
    // TODO: remove this after done with the UI
    return Observable.of(fakeData)
      .map(this.parseProviders);

    // return this.http.get(`/api/providers`, { headers: this.headers })
    //   .map(res => res.json())
    //   .map(this.parseProviders);

  }

  parseProviders(data: any): ProviderType {
    return Object.keys(data).reduce((acc, curr) => {
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
