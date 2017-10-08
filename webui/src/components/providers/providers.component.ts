import { Component, OnInit, OnDestroy } from '@angular/core';
import { ApiService } from '../../services/api.service';
import { Subscription } from 'rxjs/Subscription';

@Component({
  selector: 'providers',
  templateUrl: 'providers.component.html'
})
export class ProvidersComponent implements OnInit, OnDestroy {
  sub: Subscription;
  keys: string[];
  providers: any;
  tab: string;

  constructor(private apiService: ApiService) { }

  ngOnInit() {
    this.sub = this.apiService.fetchProviders().subscribe(data => {
      this.providers = data;
      this.keys = Object.keys(this.providers);
      this.tab = this.keys[0];

      console.log(this.providers);
    });
  }

  ngOnDestroy() {
    if (this.sub) {
      this.sub.unsubscribe();
    }
  }
}
