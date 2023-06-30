import { controller } from '@github/catalyst'

import uPlot from 'uplot';

@controller
class PlotStatElement extends HTMLElement {
    ts: number[];
    plot: uPlot;

    connectedCallback() {
        this.ts = JSON.parse(this.dataset.ts!) as number[];
        this.change(this.dataset);
    }

    change(set: DOMStringMap) {
        if (this.plot) {
            this.plot.destroy();
        }
        const y = JSON.parse(set.values!) as number[];
        this.plot = new uPlot({
            width: this.parentElement?.scrollWidth!,
            height: 240,
            series: [
                {},
                {
                    label: set.label,
                    fill: "#ffe0d8"
                },
            ],
            ms: 1
        }, [this.ts, y], this)
    }

}


if (!window.customElements.get('plot-stat')) {
    window.PlotStatElement = PlotStatElement
    window.customElements.define('plot-stat', PlotStatElement)
}