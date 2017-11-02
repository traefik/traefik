import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { HttpModule } from '@angular/http';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NgxChartsModule } from '@swimlane/ngx-charts';
import { ApiService } from './services/api.service';
import { AppComponent } from './app.component';
import { HeaderComponent } from './components/header';
import { ProvidersComponent } from './components/providers';
import { HealthComponent } from './components/health';
import { LineChartComponent } from './charts/line-chart';

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    ProvidersComponent,
    HealthComponent,
    LineChartComponent
  ],
  imports: [
    BrowserModule,
    CommonModule,
    HttpModule,
    NgxChartsModule,
    BrowserAnimationsModule,
    RouterModule.forRoot([
      { path: '', component: ProvidersComponent, pathMatch: 'full' },
      { path: 'status', component: HealthComponent }
    ])
  ],
  providers: [
    ApiService
  ],
  bootstrap: [ AppComponent ]
})
export class AppModule { }
