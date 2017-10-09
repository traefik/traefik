import { Component, ElementRef, OnInit, Input, OnChanges, SimpleChanges } from '@angular/core';
import {
  select,
  scaleTime,
  scaleLinear,
  scaleOrdinal,
  timeParse,
  line,
  curveLinear,
  extent,
  min,
  max,
  axisBottom,
  axisLeft,
  area,
  easeLinear,
  range,
  randomUniform
} from 'd3';

@Component({
  selector: 'line-chart',
  templateUrl: 'line-chart.component.html'
})
export class LineChartComponent implements OnInit {
  @Input() value: { count: number, date: Date };

  n: number;
  lineChartEl: HTMLElement;
  svg: any;
  height: number;
  width: number;
  g: any;
  line: any;
  path: any;
  area: any;
  areaPath: any;
  x: any;
  y: any;
  axisX: any;
  axisY: any;
  data: { count: number, date: number }[];
  now: number;
  duration: number;
  margin = { top: 40, right: 60, bottom: 40, left: 70 };
  parseTime: any;
  lastTime: number;
  dots: any;
  random = randomUniform(0, 20);

  constructor(private elementRef: ElementRef) { }

  ngOnInit() {
    this.lineChartEl = this.elementRef.nativeElement.querySelector('.line-chart');
    this.now = Date.now() - this.duration;

    this.render();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!this.value || !this.svg) {
      return;
    }

    this.updateData(this.value.count, this.value.date);
  }

  render() {
    this.lastTime = Date.now();
    this.n = 40;
    // this.data = range(this.n).map(() => ({ count: this.random(), date: Date.now() }));
    this.data = range(this.n).map(() => ({ count: 0, date: Date.now() }));

    this.width = this.lineChartEl.clientWidth - this.margin.left - this.margin.right
    this.height = this.lineChartEl.clientHeight - this.margin.top - this.margin.bottom;

    this.x = scaleLinear().domain([0, this.n - 1]).range([0, this.width]);
    this.y = scaleLinear().domain([0, 0.05]).range([this.height, 0]);

    this.line = line()
      .x((d: any, i: number) => this.x(i))
      .y((d: any, i: number) => this.y(d.count * 1000));

    this.area = area()
      .curve(curveLinear)
      .x((d: any, i: number) => this.x(i))
      .y1((d: any) => this.y(d.count * 1000));

    this.area.y0(this.y(0));

    this.svg = select(this.lineChartEl).append('svg')
      .attr('width', this.width + this.margin.left + this.margin.right)
      .attr('height', this.height + this.margin.top + this.margin.bottom)

    this.g = this.svg.append('g')
      .attr('transform', `translate(${this.margin.left}, ${this.margin.top})`);

    const defs = this.svg.append('defs');

    defs.append('clipPath')
      .attr('id', 'clip')
      .append('rect')
      .attr('width', this.width)
      .attr('height', this.height);

    const gradient = defs.append('linearGradient')
      .attr('id', 'line-chart-gradient')
      .attr('x1', '0%').attr('y1', '13%')
      .attr('x2', '0%').attr('y2', '100%');

    gradient.append('stop')
      .attr('stop-color', '#00AAFF')
      .attr('stop-opacity', '0.1')
      .attr('offset', 0);

    gradient.append('stop')
      .attr('stop-color', 'rgba(0,172,255,0.40)')
      .attr('stop-opacity', '0.1')
      .attr('offset', 1);

    this.axisX = this.g.append('g')
      .attr('class', 'axis axis--x')
      .attr('transform', `translate(0, ${this.height + 50})`)
      .call(axisBottom(this.x));

    this.axisY = this.g.append('g')
      .attr('class', 'axis axis--y')
      .call(axisLeft(this.y));

    this.path = this.g.append('g')
      .attr('clip-path', 'url(#clip)')
      .append('path')
      .datum(this.data)
      .attr('class', 'line')
      .attr('d', this.line);

    this.areaPath = this.g.append('path')
      .datum(this.data)
      .attr('class', 'area')
      .attr('d', this.area)
      .attr('fill', `url(${window.location.href}#line-chart-gradient)`);

    this.dots = this.g.append('g');

    this.dots.selectAll('.dot')
      .data(this.data)
      .enter().append('circle')
      .attr('class', 'dot')
      .attr('r', 3)
      .attr('cx', (d: any, i: number) => this.x(i))
      .attr('cy', (d: any, i: number) => this.y(d.count * 1000))
      .attr('fill', 'white')
      .attr('stroke', '#00AAFF')
      .attr('stroke-width', 2)
      .attr('transform', `translate(${this.x(-1)}, 0)`);

    const legend = this.svg.append('g')
      .attr('class', 'legend')
      .attr('width', 100)
      .attr('height', 100)
      .attr('transform', `translate(-${this.width / 2 - 50}, 0)`);

    legend.append('circle')
      .attr('r', 6)
      .attr('cx', this.width - 65)
      .attr('cy', 20)
      .attr('width', 10)
      .attr('height', 10)
      .style('fill', '#00AAFF');

    legend.append('text')
      .attr('x', this.width - 50)
      .attr('y', 23)
      .text('average response time (ms)');
  }

  updateData(value: number, date: Date): void {
    this.svg.selectAll('*').interrupt();
    this.duration = Date.now() - this.lastTime;
    this.lastTime = Date.now();
    
    // this.data.push({ count: this.random(), date: new Date(date).getTime() });
    this.data.push({ count: value, date: new Date(date).getTime() });
    this.y = scaleLinear().domain([0, max(this.data, (d: any) => d.count * 1000)]).range([this.height, 0]);

    this.data.shift();
    this.path
      .transition()
      .duration(0)
      .attr('d', this.line)
      .attr('transform', null)
      .transition()
      .duration(this.duration)
      .ease(easeLinear)
      .attr('transform', `translate(${this.x(-1)}, 0)`);

    this.dots.selectAll('.dot')
      .attr('cx', (d: any, i: number) => this.x(i === this.data.length - 1 ? 0 : i + 1))
      .attr('cy', (d: any, i: number) => this.y(this.data[i === this.data.length - 1 ? 0 : i + 1].count * 1000))
      .attr('transform', `translate(${this.x(0)}, 0)`);

    this.dots.selectAll('.dot')
      .transition()
      .duration(this.duration)
      .ease(easeLinear)
      .attr('transform', `translate(${this.x(-1)}, 0)`);

    this.areaPath
      .transition()
      .duration(0)
      .attr('d', this.area)
      .attr('transform', null)
      .transition()
      .duration(this.duration)
      .ease(easeLinear)
      .attr('transform', `translate(${this.x(-1)}, 0)`);

    this.axisY
      .transition()
      .duration(this.duration)
      .ease(easeLinear)
      .call(axisLeft(this.y).scale(this.y).tickSizeInner(-this.width).tickSizeOuter(0).tickPadding(25));
  }
}