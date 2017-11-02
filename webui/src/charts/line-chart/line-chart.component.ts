import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';

@Component({
  selector: 'line-chart',
  templateUrl: 'line-chart.component.html'
})
export class LineChartComponent implements OnChanges {
  @Input() value: { count: number, date: string };

  data: any[] = [];
  lineData: any[] = [];
  // options
  showXAxis = true;
  showYAxis = true;
  gradient = false;
  showLegend = true;
  showXAxisLabel = false;
  showYAxisLabel = true;
  yAxisLabel = 'Time';
  schemeType: 'linear';
  colorScheme = {
    domain: ['#0294FF']
  };
  autoScale = false;

  onSelect(event) { }

  ngOnChanges(changes: SimpleChanges) {
    if (!this.value) {
      return;
    }

    this.updateData(this.value.count, this.value.date);
  }

  updateData(count: number, date: string): void {
    this.lineData.push({
      name: new Date(),
      value: count * 1000,
    });

    this.data = [{ "name": "Avg. Response Time", "series": this.lineData }];
  }
}
