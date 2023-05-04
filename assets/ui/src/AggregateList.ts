import { controller, targets } from '@github/catalyst'

@controller
class MainAggregateListElement extends HTMLElement {
    @targets items: HTMLElement[];

    connectedCallback() { }

    select(event: Event) {
        const button = event.currentTarget as HTMLButtonElement
        for (const item of this.items) {
            item.classList.remove('propertySelected')
        }
        button.classList.add('propertySelected')
    }
}

declare global {
    interface Window {
        MainAggregateListElement: typeof MainAggregateListElement
    }
}
if (!window.customElements.get('main-aggregate-list')) {
    window.MainAggregateListElement = MainAggregateListElement
    window.customElements.define('main-aggregate-list', MainAggregateListElement)
}