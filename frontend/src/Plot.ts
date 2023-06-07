import { controller, target, targets, attr } from '@github/catalyst'

import { Options, AlignedData } from "uplot";
import uPlot from 'uplot';

@controller
class PlotStatElement extends HTMLElement {
    @attr width: number;
    @attr height: number;
    @attr label: string;
    @attr ts: string;
    @attr value: string;
    plot: uPlot;

    connectedCallback() {
        this.plot = new uPlot(this.#options(), this.#data(), this);
    }

    #options(): Options {
        uPlot.clipGaps
        return {
            width: this.width,
            height: this.height,
            series: [
                {},
                {
                    label: this.label,
                    fill: "#ffe0d8"
                },
            ],
            ms: 1
        }
    }

    #data(): AlignedData {
        const x = JSON.parse(this.ts) as number[];
        const y = JSON.parse(this.value) as number[];
        return [x, y];
    }

}

function shortNumber(value: number): string {
    let p = Math.floor(Math.log10(Math.abs(value)));
    if (p <= 2) return value.toString();
    let l = Math.floor(p / 3);
    let shortened =
        Math.pow(10, p - l * 3) * +(value / Math.pow(10, p)).toFixed(1);
    return Math.round(shortened * 100) / 100 + " " + ["", "K", "M", "B", "T"][l];
}

declare global {
    interface Window {
        PlotStatElement: typeof PlotStatElement
    }
}



if (!window.customElements.get('plot-stat')) {
    window.PlotStatElement = PlotStatElement
    window.customElements.define('plot-stat', PlotStatElement)
}

