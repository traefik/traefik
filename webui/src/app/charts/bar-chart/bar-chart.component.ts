import { Component, ElementRef, Input, OnChanges, OnInit, SimpleChanges } from '@angular/core';
import { axisBottom, axisLeft, max, scaleBand, scaleLinear, select } from 'd3';
import { format } from 'd3-format';
import * as _ from 'lodash';
import { WindowService } from '../../services/window.service';

@Component({
  selector: 'app-bar-chart',
  templateUrl: './bar-chart.component.html'
})
export class BarChartComponent implements OnInit, OnChanges {
  @Input() value: any;

  barChartEl: HTMLElement;
  svg: any;
  x: any;
  y: any;
  g: any;
  width: number;
  height: number;
  margin = {top: 40, right: 40, bottom: 40, left: 40};
  loading: boolean;
  data: any[];
  previousData: any[];

  constructor(public elementRef: ElementRef, public windowService: WindowService) {
    this.loading = true;
  }

  ngOnInit() {
    this.barChartEl = this.elementRef.nativeElement.querySelector('.bar-chart');
    this.setup();
    setTimeout(() => this.loading = false, 1000);

    this.windowService.resize.subscribe(w => this.draw());
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!this.value || !this.svg) {
      return;
    }

    if (!_.isEqual(this.previousData, this.value)) {
      this.previousData = _.cloneDeep(this.value);
      this.data = this.value;

      this.draw();
    }
  }

  setup(): void {
    this.width = this.barChartEl.clientWidth - this.margin.left - this.margin.right;
    this.height = this.barChartEl.clientHeight - this.margin.top - this.margin.bottom;

    this.svg = select(this.barChartEl)
      .append('svg')
      .attr('width', this.width + this.margin.left + this.margin.right)
      .attr('height', this.height + this.margin.top + this.margin.bottom);

    this.g = this.svg.append('g')
      .attr('transform', `translate(${this.margin.left}, ${this.margin.top})`);

    this.x = scaleBand().padding(0.05);
    this.y = scaleLinear();

    this.g.append('g')
      .attr('class', 'axis axis--x');

    this.g.append('g')
      .attr('class', 'axis axis--y');
  }

  draw(): void {
    if (this.barChartEl.clientWidth === 0 || this.barChartEl.clientHeight === 0) {
      this.previousData = [];
    } else {
      this.width = this.barChartEl.clientWidth - this.margin.left - this.margin.right;
      this.height = this.barChartEl.clientHeight - this.margin.top - this.margin.bottom;
    }

    this.x.domain(this.data.map((d: any) => d.code));
    this.y.domain([0, max(this.data, (d: any) => d.count)]);


    this.svg
      .attr('width', this.width + this.margin.left + this.margin.right)
      .attr('height', this.height + this.margin.top + this.margin.bottom);

    this.x.rangeRound([0, this.width]);
    this.y.rangeRound([this.height, 0]);

    this.g.select('.axis--x')
      .attr('transform', `translate(0, ${this.height})`)
      .call(axisBottom(this.x));

    this.g.select('.axis--y')
      .call(axisLeft(this.y).tickFormat(format('~s')).tickSize(-this.width));

    // Clean previous graph
    this.g.selectAll('.bar').remove();

    const bars = this.g.selectAll('.bar').data(this.data);

    bars.enter()
      .append('rect')
      .attr('class', 'bar')
      .style('fill', (d: any) => 'hsl(' + Math.floor(((d.code - 100) * 310 / 427) + 50) + ', 50%, 50%)')
      .attr('x', (d: any) => this.x(d.code))
      .attr('y', (d: any) => this.y(d.count))
      .attr('width', this.x.bandwidth())
      .attr('height', (d: any) => (this.height - this.y(d.count)) < 0 ? 0 : this.height - this.y(d.count));

    bars.exit().remove();
  }

}
