
import { controller, target, targets } from '@github/catalyst'

@controller
export class StatPanelsElement extends HTMLElement {
    @targets items: PropPanelsElement[];

    change(event: Event) {
        const e = event as CustomEvent;
        const target = e.detail.relatedTarget as HTMLElement;
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
    @target active: HTMLElement;

    change(event: Event) {
        const e = event as CustomEvent;
        const target = e.detail.relatedTarget as HTMLElement;
        this.active = target;
    }

    select(metric: string) {
        for (let e of this.active.children!) {
            const el = e as HTMLElement;
            if (el.dataset.metric == metric) {
                el.hidden = false
            } else {
                el.hidden = true;
            }
        }
    }
}

declare global {
    interface Window {
        StatPanelsElement: typeof StatPanelsElement
        PropPanelsElement: typeof PropPanelsElement
    }
}

if (!window.customElements.get('stat-panels')) {
    window.StatPanelsElement = StatPanelsElement
    window.customElements.define('stat-panels', StatPanelsElement)
}
if (!window.customElements.get('prop-panels')) {
    window.PropPanelsElement = PropPanelsElement
    window.customElements.define('prop-panels', PropPanelsElement)
}