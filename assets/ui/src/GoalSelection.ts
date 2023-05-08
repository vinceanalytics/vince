import { controller, target } from '@github/catalyst'

@controller
class GoalSelectionElement extends HTMLElement {
    @target event: HTMLButtonElement;
    @target event_fields: HTMLElement;
    @target path: HTMLButtonElement;
    @target path_fields: HTMLElement;

    selectEvent(e: Event) {
        this.event.classList.add('propertySelected')
        this.path.classList.remove('propertySelected')
        this.event_fields.classList.remove('d-none')
        this.path_fields.classList.add('d-none')
    }
    selectPath(e: Event) {
        this.path.classList.add('propertySelected')
        this.event.classList.remove('propertySelected')
        this.event_fields.classList.add('d-none')
        this.path_fields.classList.remove('d-none')
    }
}

declare global {
    interface Window {
        GoalSelectionElement: typeof GoalSelectionElement
    }
}

if (!window.customElements.get('goal-selection')) {
    window.GoalSelectionElement = GoalSelectionElement
    window.customElements.define('goal-selection', GoalSelectionElement)
}