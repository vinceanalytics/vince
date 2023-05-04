import { controller, targets } from '@github/catalyst'

@controller
class SelectionListElement extends HTMLElement {
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
        SelectionListElement: typeof SelectionListElement
    }
}
if (!window.customElements.get('selection-list')) {
    window.SelectionListElement = SelectionListElement
    window.customElements.define('selection-list', SelectionListElement)
}