import { controller, target, targets, attr } from '@github/catalyst'

import { Options, AlignedData } from "uplot";
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

@controller
export class StatPanelsElement extends HTMLElement {
    @targets items: PropPanelsElement[];
    @target plot: PlotStatElement;

    change(event: Event) {
        const e = event as CustomEvent;
        const target = e.detail.relatedTarget as HTMLElement;
        this.plot.change(target.dataset);
        this.select(target.dataset.metric!);
    }

    select(metric: string) {
        this.items.forEach((e) => {
            e.select(metric)
        })
    }
}

@controller
export class PropPanelsElement extends HTMLElement {
    @targets items: HTMLElement[];
    @targets panels: HTMLElement[];
    @attr active: string;
    @attr metric: string;

    change(event: Event) {
        const e = event as CustomEvent;
        const target = e.detail.relatedTarget as HTMLElement;
        this.active = target.dataset.prop!;
        this.select(this.metric);
    }

    select(metric: string) {
        this.metric = metric;
        this.panels.forEach(element => {
            if (element.dataset.prop == this.active) {
                for (let e of element.children!) {
                    const el = e as HTMLElement;
                    if (el.dataset.metric == metric) {
                        el.hidden = false
                    } else {
                        el.hidden = true;
                    }
                }
            }
        });
    }
}


declare global {
    interface Window {
        PlotStatElement: typeof PlotStatElement
        StatPanelsElement: typeof StatPanelsElement
        PropPanelsElement: typeof PropPanelsElement
    }
}



if (!window.customElements.get('plot-stat')) {
    window.PlotStatElement = PlotStatElement
    window.customElements.define('plot-stat', PlotStatElement)
}
if (!window.customElements.get('stat-panels')) {
    window.StatPanelsElement = StatPanelsElement
    window.customElements.define('stat-panels', StatPanelsElement)
}
if (!window.customElements.get('prop-panels')) {
    window.PropPanelsElement = PropPanelsElement
    window.customElements.define('prop-panels', PropPanelsElement)
}
