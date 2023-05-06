import { controller, targets } from '@github/catalyst'
import { Aggregate, Entry, EntryMap } from './Data';

@controller
class VinceStatsElement extends HTMLElement {
    @targets all_box: HTMLElement[];
    @targets pages_box: HTMLElement[];
    @targets referrer_box: HTMLElement[];
    @targets countries_box: HTMLElement[];
    @targets os_box: HTMLElement[];

    state: State = {};

    connectedCallback() { }

    selectAllBox(event: Event) {
        this.#select(event, this.all_box);
    }

    selectPagesBox(event: Event) {
        this.#select(event, this.pages_box);
    }

    referrerBox(event: Event) {
        this.#select(event, this.referrer_box);
    }

    countriesBox(event: Event) {
        this.#select(event, this.countries_box);
    }

    osBox(event: Event) {
        this.#select(event, this.os_box);
    }

    #select(event: Event, elements: HTMLElement[]) {
        const button = event.currentTarget as HTMLButtonElement
        for (const item of elements) {
            item.classList.remove('propertySelected')
        }
        button.classList.add('propertySelected')
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
    timestamps?: number[];
    all_box?: Aggregate;
    all_box_plot?: Entry;
    pages_box?: EntryMap;
    referrer_box?: EntryMap;
    countries_box?: EntryMap;
    os_box?: EntryMap;
    goals_box?: EntryMap;
    events_box?: EntryMap;
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