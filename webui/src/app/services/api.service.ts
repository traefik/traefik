import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/map';

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
    return this.http.get(`/api/version`, { headers: this.headers });
  }

  fetchHealthStatus(): Observable<any> {
    return this.http.get(`/health`, { headers: this.headers });
  }

  fetchProviders(): Observable<any> {
    return this.http.get(`/api/providers`, { headers: this.headers })
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
