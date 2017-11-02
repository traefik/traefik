import { Component, Input, OnInit, ElementRef, OnChanges, SimpleChanges } from '@angular/core';
import { range, scaleTime, scaleLinear, max, curveLinear, line, easeLinear, select, axisLeft, axisBottom, timeSecond, timeFormat } from 'd3';

@Component({
  selector: 'line-chart',
  templateUrl: 'line-chart.component.html'
})
export class LineChartComponent implements OnChanges, OnInit {
  @Input() value: { count: number, date: string };

  lineChartEl: HTMLElement;
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
  height: number;
  width: number;
  margin = { top: 40, right: 40, bottom: 40, left: 40 };
  loading = true;

  constructor(private elementRef: ElementRef) { }

  ngOnInit() {
    this.lineChartEl = this.elementRef.nativeElement.querySelector('.line-chart');
    this.limit = 20;
    this.duration = 3000;
    this.now = new Date(Date.now() - this.duration);

    this.options = {
      title: '',
      color: '#3A84C5'
    };

    this.render();

    setTimeout(() => this.loading = false, 4000);
  }

  render() {
    this.width = this.lineChartEl.clientWidth - this.margin.left - this.margin.right;
    this.height = this.lineChartEl.clientHeight - this.margin.top - this.margin.bottom;

    this.svg = select(this.lineChartEl).append('svg')
      .attr('width', this.width + this.margin.left + this.margin.right)
      .attr('height', this.height + this.margin.top + this.margin.bottom)
      .append('g')
      .attr('transform', `translate(${this.margin.left}, ${this.margin.top})`);

    this.data = range(this.limit).map(i => 0);

    this.x = scaleTime().range([0, this.width]);
    this.y = scaleLinear().range([this.height, 0]);

    this.x.domain([<any>this.now - (this.limit - 2), <any>this.now - this.duration]);
    this.y.domain([-0.005, 0.5]);

    this.line = line()
      .x((d: any, i: number) => this.x(<any>this.now - (this.limit - 1 - i) * this.duration))
      .y((d: any) => this.y(d))
      .curve(curveLinear);

    this.svg.append('defs').append('clipPath')
      .attr('id', 'clip')
      .append('rect')
      .attr('width', this.width)
      .attr('height', this.height);

    this.xAxis = this.svg.append('g')
      .attr('class', 'x axis')
      .attr('transform', `translate(0, ${this.height})`)
      .call(axisBottom(this.x).tickSize(-this.height));

    this.svg.append('g')
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

    this.updateData(this.value.count)
  }

  updateData = (value: number) => {
    this.data.push(value * 10000);
    this.now = new Date();

    this.x.domain([<any>this.now - (this.limit - 2) * this.duration, <any>this.now - this.duration]);

    this.xAxis
      .transition()
      .duration(this.duration)
      .ease(easeLinear)
      .call(axisBottom(this.x).tickSize(-this.height).ticks(timeSecond, 5).tickFormat(timeFormat('%H:%M:%S')));

    this.path
      .transition()
      .duration(0)
      .attr('d', this.line(this.data))
      .attr('transform', null)
      .transition()
      .duration(this.duration)
      .ease(easeLinear)
      .attr('transform', `translate(${this.x(<any>this.now - (this.limit - 1) * this.duration)})`);

    this.data.shift();
  }
}


// import { Component, Input, OnInit, ElementRef, OnChanges, SimpleChanges } from '@angular/core';
// import { scaleLinear, scaleTime, timeSeconds, randomNormal, select, line, axisBottom, axisLeft, easeLinear, range, max, min, timeSecond, timeFormat } from 'd3';

// @Component({
//   selector: 'line-chart',
//   templateUrl: 'line-chart.component.html'
// })
// export class LineChartComponent implements OnChanges, OnInit {
//   @Input() value: { count: number, date: string };

//   el: HTMLElement;
//   data: any[];
//   n = 40;
//   random = randomNormal(4, .2);
//   i = 0;
//   margin = { top: 10, right: 10, bottom: 20, left: 40 };
//   width: number;
//   height: number;
//   axisPadding = 30;
//   x: any;
//   xAxisScale: any;
//   y: any;
//   yAxisScale: any;
//   line: any;
//   svg: any;
//   xAxis: any;
//   path: any;
//   updating = false;

//   constructor(private elementRef: ElementRef) { }

//   ngOnInit() {
//     this.el = this.elementRef.nativeElement.querySelector('.line-chart');

//     this.data = range(this.n).map((x, i) => {
//       const time = new Date().getTime() - (i * 1000);
//       return { value: 0, time: time };
//     }).reverse();
//     this.width = this.el.clientWidth - this.margin.left - this.margin.right;
//     this.height = this.el.clientHeight - this.margin.top - this.margin.bottom;

//     this.x = scaleLinear()
//       .domain([min(this.data, (d: any) => d.time), max(this.data, (d: any) => d.time)])
//       .range([0, this.width]);

//     this.xAxisScale = scaleTime()
//       .domain([min(this.data, (d: any) => d.time), max(this.data, (d: any) => d.time)])
//       .range([0, this.width]);

//     this.y = scaleLinear()
//       .domain([-1, 1])
//       .range([this.height - this.axisPadding, 0]);

//     this.line = line()
//       .x((d: any, i: number) => this.x(d.time))
//       .y((d: any, i: number) => this.y(d.value));

    // this.svg = select(this.el).append('svg')
    //   .attr('width', this.width + this.margin.left + this.margin.right)
    //   .attr('height', this.height + this.margin.top + this.margin.bottom)
    //   .append('g')
    //   .attr('transform', `translate(${this.margin.left}, ${this.margin.top})`);

//     this.svg.append('defs').append('clipPath')
//       .attr('id', 'clip')
//       .append('rect')
//       .attr('width', this.width)
//       .attr('height', this.height);

//     this.xAxis = this.svg.append('g')
//       .attr('class', 'x axis')
//       .attr('transform', `translate(0, ${this.height})`)
//       .call(axisBottom(this.x).tickSize(-this.height));

//     this.svg.append('g')
//       .attr('class', 'y axis')
//       .call(axisLeft(this.y));

//     this.path = this.svg.append('g')
//       .attr('clip-path', 'url(#clip)')
//       .append('path')
//       .data([this.data])
//       .attr('class', 'line');
//   }

//   ngOnChanges(changes: SimpleChanges) {
//     if (!this.value || !this.svg) {
//       return;
//     }

//     if (!this.updating) {
//       this.updateData(this.value.count, this.value.date);
//       this.updating = true;
//     }
//   }

//   updateData(count: number, date: string): void {
//     let self = this;
//     this.i++;
//     this.data.push({ value: count, time: new Date().getTime() });

//     this.xAxisScale = scaleTime()
//       .domain([min(this.data, (d: any) => d.time), max(this.data, (d: any) => d.time)])
//       .range([0, this.width]);

//     this.xAxis
//       .transition()
//       .duration(1000)
//       .ease(easeLinear)
//       .call(axisBottom(this.xAxisScale).tickSize(-this.height).ticks(timeSecond, 10).tickFormat(timeFormat('%H:%M:%S')));

//     this.path
//       .attr('d', this.line)
//       .attr('transform', null)
//       .transition()
//       .duration(1000)
//       .ease(easeLinear)
//       .attr('transform', `translate(${this.x(new Date(0))})`)
//       .on('end', () => self.updateData(this.value.count, this.value.date));

//     this.data.shift();
//   }
// }
