import { Component, OnInit } from '@angular/core';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-header',
  templateUrl: 'header.component.html'
})
export class HeaderComponent implements OnInit {
  version: string;
  codename: string;
  releaseLink: string;
  burger: boolean;

  constructor(private apiService: ApiService) { }

  ngOnInit() {
    this.apiService.fetchVersion()
      .subscribe(data => {
        this.version = data.Version;
        this.codename = data.Codename;
        this.releaseLink = 'https://github.com/containous/traefik/tree/' + data.Version;
      });
  }
}
