import { Component, ElementRef, Input, OnChanges, OnInit, SimpleChanges } from '@angular/core';
import {
  axisBottom,
  axisLeft,
  curveLinear,
  easeLinear,
  line,
  max,
  min,
  range,
  scaleLinear,
  scaleTime,
  select,
  timeFormat,
  timeSecond
} from 'd3';
import { WindowService } from '../../services/window.service';

@Component({
  selector: 'app-line-chart',
  templateUrl: 'line-chart.component.html'
})
export class LineChartComponent implements OnChanges, OnInit {
  @Input() value: { count: number, date: string };

  firstDisplay: boolean;
  dirty: boolean;
  lineChartEl: HTMLElement;
  loadingEl: HTMLElement;
  svg: any;
  g: any;
  line: any;
  path: any;
  x: any;
  y: any;
  data: number[];
  now: Date;
  duration: number;
  limit: number;
  options: any;
  xAxis: any;
  yAxis: any;
  height: number;
  width: number;
  margin = {top: 40, right: 40, bottom: 60, left: 60};
  loading = true;

  constructor(private elementRef: ElementRef, public windowService: WindowService) { }

  ngOnInit() {
    this.lineChartEl = this.elementRef.nativeElement.querySelector('.line-chart');
    this.loadingEl = this.elementRef.nativeElement.querySelector('.line-chart-loading');
    this.limit = 40;

    // related to the Observable.timer(0, 3000) in health component
    this.duration = 3000;

    this.now = new Date(Date.now() - this.duration);

    this.options = {
      title: '',
      color: '#3A84C5'
    };

    this.firstDisplay = true;
    this.render();

    this.windowService.resize.subscribe(w => {
      if (this.svg) {
        this.dirty = true;
        this.loading = true;
        this.render();
      }
    });
  }

  render() {
    // When the lineChartEl is not displayed (is-hidden), width and length are equal to 0.
    let elt;
    if (this.lineChartEl.clientWidth === 0 || this.lineChartEl.clientHeight === 0) {
      elt = this.loadingEl;
    } else {
      elt = this.lineChartEl;
    }
    this.width = elt.clientWidth - this.margin.left - this.margin.right;
    this.height = elt.clientHeight - this.margin.top - this.margin.bottom;


    const el = this.lineChartEl.querySelector('svg');
    if (el) {
      el.parentNode.removeChild(el);
    }

    this.svg = select(this.lineChartEl)
      .append('svg')
      .attr('width', this.width + this.margin.left + this.margin.right)
      .attr('height', this.height + this.margin.top + this.margin.bottom)
      .append('g')
      .attr('transform', `translate(${this.margin.left}, ${this.margin.top})`);

    if (!this.data) {
      this.data = range(this.limit).map(i => 0);
    }

    this.x = scaleTime().range([0, this.width - 10]);
    this.y = scaleLinear().range([this.height, 0]);

    this.x.domain([<any>this.now - (this.limit - 2), <any>this.now - this.duration]);
    this.y.domain([0, max(this.data, (d: any) => d)]);

    this.line = line()
      .x((d: any, i: number) => this.x(<any>this.now - (this.limit - 1 - i) * this.duration))
      .y((d: any) => this.y(d))
      .curve(curveLinear);

    this.svg
      .append('defs')
      .append('clipPath')
      .attr('id', 'clip')
      .append('rect')
      .attr('width', this.width)
      .attr('height', this.height);

    this.xAxis = this.svg.append('g')
      .attr('class', 'x axis')
      .attr('transform', `translate(0, ${this.height})`)
      .call(axisBottom(this.x).tickSize(-this.height).ticks(timeSecond, 5).tickFormat(timeFormat('%H:%M:%S')));

    this.yAxis = this.svg.append('g')
      .attr('class', 'y axis')
      .call(axisLeft(this.y).tickSize(-this.width));

    this.path = this.svg.append('g')
      .attr('clip-path', 'url(#clip)')
      .append('path')
      .data([this.data])
      .attr('class', 'line');
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!this.value || !this.svg) {
      return;
    }

    this.updateData(this.value.count);
  }

  updateData(value: number) {
    this.data.push(value * 1000000);
    this.now = new Date();

    this.x.domain([<any>this.now - (this.limit - 2) * this.duration, <any>this.now - this.duration]);
    const minv = min(this.data, (d: any) => d) > 0 ? min(this.data, (d: any) => d) - 4 : 0;
    const maxv = max(this.data, (d: any) => d) + 4;
    this.y.domain([minv, maxv]);

    this.xAxis
      .transition()
      .duration(this.firstDisplay || this.dirty ? 0 : this.duration)
      .ease(easeLinear)
      .call(axisBottom(this.x).tickSize(-this.height).ticks(timeSecond, 5).tickFormat(timeFormat('%H:%M:%S')));

    this.xAxis
      .transition()
      .duration(0)
      .selectAll('text')
      .style('text-anchor', 'end')
      .attr('dx', '-.8em')
      .attr('dy', '.15em')
      .attr('transform', 'rotate(-65)');

    this.yAxis
      .transition()
      .duration(500)
      .ease(easeLinear)
      .call(axisLeft(this.y).tickSize(-this.width));

    this.path
      .transition()
      .duration(0)
      .attr('d', this.line(this.data))
      .attr('transform', null)
      .transition()
      .duration(this.duration)
      .ease(easeLinear)
      .attr('transform', `translate(${this.x(<any>this.now - (this.limit - 1) * this.duration)})`);

    this.firstDisplay = false;
    this.dirty = false;

    if (this.loading) {
      this.loading = false;
    }

    this.data.shift();
  }
}
