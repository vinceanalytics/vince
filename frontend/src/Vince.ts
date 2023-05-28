import { controller, targets, target } from '@github/catalyst'

@controller
class VinceStatsElement extends HTMLElement {
    @targets periods: HTMLElement[];
    @target period_label: HTMLElement;

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
        this.period_label.innerText = a?.dataset.name!;
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