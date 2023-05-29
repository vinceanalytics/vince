import { controller, targets, target } from '@github/catalyst'

@controller
class VinceStatsElement extends HTMLElement {
    @targets periods: HTMLElement[];
    @target period_label: HTMLElement;
    @target period_range_from_label: HTMLElement;
    @target period_range_to_label: HTMLElement;

    @targets metrics: HTMLElement[];
    @targets props: HTMLElement[];


    connectedCallback() { }


    changePeriod(e: Event) {
        const target = e.target as HTMLElement;
        this.periods.forEach((period) => {
            period.querySelector(".select-menu-item-icon")?.
                classList.add("d-none")
        });
        const a = target.parentElement;
        a?.
            querySelector(".select-menu-item-icon")?.
            classList.remove("d-none")
        const ds = a?.dataset;
        this.period_label.innerText = ds?.name!;
        this.period_range_from_label.innerText = ds?.fromLabel!;
        this.period_range_to_label.innerText = ds?.toLabel!;
    }

    changeMetrics(e: Event) {
        const target = e.target as HTMLElement;
        this.metrics.forEach((metric) => {
            metric.parentElement?.
                classList.remove("navigation-focus")
        });
        const a = target.parentElement;
        a?.classList.add("navigation-focus");
    }

    changeProps(e: Event) {
        const target = e.target as HTMLElement;
        this.props.forEach((prop) => {
            prop.
                classList.remove("activeProps")
            prop.
                classList.add("border-0")
        });
        target.classList.add("activeProps");
        target.classList.remove("border-0");
    }


    #shortNumber(value: number): string {
        let p = Math.floor(Math.log10(Math.abs(value)));
        if (p <= 2) return value.toString();
        let l = Math.floor(p / 3);
        let shortened =
            Math.pow(10, p - l * 3) * +(value / Math.pow(10, p)).toFixed(1);
        return Math.round(shortened * 100) / 100 + " " + ["", "K", "M", "B", "T"][l];
    }
}

interface State {

}


declare global {
    interface Window {
        VinceStatsElement: typeof VinceStatsElement
    }
}

if (!window.customElements.get('vince-stats')) {
    window.VinceStatsElement = VinceStatsElement
    window.customElements.define('vince-stats', VinceStatsElement)
}