import { controller, attr } from '@github/catalyst'

import uPlot from 'uplot';

@controller
class PlotStatElement extends HTMLElement {
    @attr timestamps: string;
    @attr values: string;
    @attr label: string;

    x: number[];
    y: number[];
    plot: uPlot;
    obs: ResizeObserver;

    connectedCallback() {
        this.x = JSON.parse(this.timestamps) as number[];
        this.y = JSON.parse(this.values) as number[];
        this.obs = new ResizeObserver(() => this.change())
        this.obs.observe(this.parentElement!)
        this.change();
    }

    disconnectedCallback() {
        this.plot.destroy()
        this.obs.unobserve(this.parentElement!)
    }

    change() {
        if (this.plot) {
            this.plot.destroy();
        }
        this.plot = new uPlot({
            width: this.parentElement?.scrollWidth!,
            height: 240,
            series: [
                {},
                {
                    label: this.label,
                    fill: "#ffe0d8"
                },
            ],
            ms: 1
        }, [this.x, this.y], this)
    }

}


if (!window.customElements.get('plot-stat')) {
    //@ts-ignore
    window.PlotStatElement = PlotStatElement
    window.customElements.define('plot-stat', PlotStatElement)
}

function plot(id, label, x, y) {
    let el = document.getElementById(id)!
    new uPlot({ width: el.parentElement?.scrollWidth!, height: 240, series: [{}, { label: label, fill: "#ffe0d8" }], ms: 1 }, [x, y], el!);;
}

window.plot = plot;