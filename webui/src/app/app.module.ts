import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { ApiService } from './services/api.service';
import { WindowService } from './services/window.service';
import { AppComponent } from './app.component';
import { HeaderComponent } from './components/header/header.component';
import { ProvidersComponent } from './components/providers/providers.component';
import { HealthComponent } from './components/health/health.component';
import { LineChartComponent } from './charts/line-chart/line-chart.component';
import { BarChartComponent } from './charts/bar-chart/bar-chart.component';
import { KeysPipe } from './pipes/keys.pipe';
import { FrontendFilterPipe } from './pipes/frontend.filter.pipe';
import { BackendFilterPipe } from './pipes/backend.filter.pipe';

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    ProvidersComponent,
    HealthComponent,
    LineChartComponent,
    BarChartComponent,
    KeysPipe,
    FrontendFilterPipe,
    BackendFilterPipe
  ],
  imports: [
    BrowserModule,
    CommonModule,
    HttpClientModule,
    FormsModule,
    RouterModule.forRoot([
      { path: '', component: ProvidersComponent, pathMatch: 'full' },
      { path: 'status', component: HealthComponent }
    ])
  ],
  providers: [
    ApiService,
    WindowService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
