import {
  HttpClient,
  HttpErrorResponse,
  HttpHeaders
} from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, EMPTY, of } from 'rxjs';
import { catchError, map, retry } from 'rxjs/operators';
import { RawHealth } from '../components/health/health.component';

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
    return this.http.get('../api/version', { headers: this.headers }).pipe(
      retry(4),
      catchError((err: HttpErrorResponse) => {
        console.error(
          `[version] returned code ${err.status}, body was: ${err.error}`
        );
        return EMPTY;
      })
    );
  }

  fetchHealthStatus() {
    return this.http
      .get<RawHealth>('../health', { headers: this.headers })
      .pipe(
        retry(2),
        catchError((err: HttpErrorResponse) => {
          console.error(
            `[health] returned code ${err.status}, body was: ${err.error}`
          );
          return EMPTY;
        })
      );
  }

  fetchProviders(): Observable<any> {
    return this.http.get('../api/providers', { headers: this.headers }).pipe(
      retry(2),
      catchError((err: HttpErrorResponse) => {
        console.error(
          `[providers] returned code ${err.status}, body was: ${err.error}`
        );
        return of<any>({});
      }),
      map((data: any): ProviderType => this.parseProviders(data))
    );
  }

  parseProviders(data: any): ProviderType {
    return Object.keys(data)
      .filter(value => value !== 'acme' && value !== 'ACME')
      .reduce((acc, curr) => {
        acc[curr] = {};

        acc[curr].frontends = this.toArray(data[curr].frontends, 'id').map(
          frontend => {
            frontend.routes = this.toArray(frontend.routes, 'id');
            frontend.errors = this.toArray(frontend.errors, 'id');
            if (frontend.headers) {
              frontend.headers.customRequestHeaders = this.toHeaderArray(
                frontend.headers.customRequestHeaders
              );
              frontend.headers.customResponseHeaders = this.toHeaderArray(
                frontend.headers.customResponseHeaders
              );
              frontend.headers.sslProxyHeaders = this.toHeaderArray(
                frontend.headers.sslProxyHeaders
              );
            }
            if (frontend.ratelimit && frontend.ratelimit.rateset) {
              frontend.ratelimit.rateset = this.toArray(
                frontend.ratelimit.rateset,
                'id'
              );
            }
            return frontend;
          }
        );

        acc[curr].backends = this.toArray(data[curr].backends, 'id').map(
          backend => {
            backend.servers = this.toArray(backend.servers, 'id');
            return backend;
          }
        );

        return acc;
      }, {});
  }

  toHeaderArray(data: any): any[] {
    return Object.keys(data || {}).map(key => ({
      name: key,
      value: data[key]
    }));
  }

  toArray(data: any, fieldKeyName: string): any[] {
    return Object.keys(data || {}).map(key => {
      data[key][fieldKeyName] = key;
      return data[key];
    });
  }
}
