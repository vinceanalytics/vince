import { controller, target, targets, attr } from '@github/catalyst'


@controller
class MetricValueElement extends HTMLElement {
    @attr visitors = 0;
    @attr views = 0;
    @attr events = 0;
    @attr visits = 0;
    @attr bounceRates = 0;
    @attr visitDurations = 0;
    @attr viewsPerVisits = 0;

    @target visitor: HTMLElement;
    @target view: HTMLElement;
    @target event: HTMLElement;
    @target visit: HTMLElement;
    @target bounceRate: HTMLElement;
    @target visitDuration: HTMLElement;
    @target viewsPerVisit: HTMLElement;

    @targets items: HTMLElement[];

    attributeChangedCallback() {
        if (this.visitor) {
            this.visitor.textContent = `${this.visitors}`
        }
        if (this.view) {
            this.view.textContent = `${this.views}`
        }
        if (this.event) {
            this.event.textContent = `${this.events}`
        }
        if (this.visit) {
            this.visit.textContent = `${this.visits}`
        }
        if (this.bounceRate) {
            this.bounceRate.textContent = `${this.bounceRates}`
        }
        if (this.visitDuration) {
            this.visitDuration.textContent = `${this.visitDurations}`
        }
        if (this.viewsPerVisit) {
            this.viewsPerVisit.textContent = `${this.viewsPerVisits}`
        }
    }

    selectMetric(e: Event) {
        this.items.forEach((m) => {
            m.classList.remove("navigation-focus")
        })
        const a = e.target as HTMLElement;
        if (a.classList.contains("Box-row--focus-blue")) {
            a.classList.add("navigation-focus");
            a.dispatchEvent(new CustomEvent("change"))
        } else {
            // child elements were clicked
            const p = a.parentElement!
            p.classList.add("navigation-focus");
            p.dispatchEvent(new CustomEvent("change"))
        }
    }
}


declare global {
    interface Window {
        MetricValueElement: typeof MetricValueElement
    }
}



if (!window.customElements.get('metric-value')) {
    window.MetricValueElement = MetricValueElement
    window.customElements.define('metric-value', MetricValueElement)
}

