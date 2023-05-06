import { controller, targets } from '@github/catalyst'

@controller
class VinceStatsElement extends HTMLElement {
    @targets all_box: HTMLElement[];
    @targets pages_box: HTMLElement[];
    @targets referrer_box: HTMLElement[];
    @targets countries_box: HTMLElement[];
    @targets os_box: HTMLElement[];

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