import { CommonModule } from '@angular/common';
import { HttpClientModule } from '@angular/common/http';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { AppComponent } from './app.component';
import { BarChartComponent } from './charts/bar-chart/bar-chart.component';
import { LineChartComponent } from './charts/line-chart/line-chart.component';
import { HeaderComponent } from './components/header/header.component';
import { HealthComponent } from './components/health/health.component';
import { ProvidersComponent } from './components/providers/providers.component';
import { LetDirective } from './directives/let.directive';
import { BackendFilterPipe } from './pipes/backend.filter.pipe';
import { FrontendFilterPipe } from './pipes/frontend.filter.pipe';
import { HumanReadableFilterPipe } from './pipes/humanreadable.filter.pipe';
import { KeysPipe } from './pipes/keys.pipe';
import { ApiService } from './services/api.service';
import { WindowService } from './services/window.service';

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
    BackendFilterPipe,
    HumanReadableFilterPipe,
    LetDirective
  ],
  imports: [
    BrowserModule,
    CommonModule,
    HttpClientModule,
    FormsModule,
    RouterModule.forRoot([
      {path: '', component: ProvidersComponent, pathMatch: 'full'},
      {path: 'status', component: HealthComponent}
    ])
  ],
  providers: [
    ApiService,
    WindowService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
