import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { WindowService } from '../../services/window.service';
import { BarChartComponent } from './bar-chart.component';

describe('BarChartComponent', () => {
  let component: BarChartComponent;
  let fixture: ComponentFixture<BarChartComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [BarChartComponent],
      providers: [WindowService]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(BarChartComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initially go to loading state', () => {
    expect(component.loading).toBeTruthy();
  });
});
