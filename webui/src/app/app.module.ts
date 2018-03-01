import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { ApiService } from './services/api.service';
import { AppComponent } from './app.component';
import { HeaderComponent } from './components/header/header.component';
import { ProvidersComponent } from './components/providers/providers.component';
import { HealthComponent } from './components/health/health.component';
import { LineChartComponent } from './charts/line-chart';
import { KeysPipe } from './pipes/keys.pipe';

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    ProvidersComponent,
    HealthComponent,
    LineChartComponent,
    KeysPipe
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
    ApiService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
