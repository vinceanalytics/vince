
import { controller, target } from '@github/catalyst'

@controller
export class SendFormElement extends HTMLElement {
    @target link: HTMLElement;

    send(event: Event) {
        event.preventDefault();
        event.stopImmediatePropagation();
        let form = document.createElement('form');
        let submit = document.createElement("input");
        form.method = this.link.dataset.method!;
        form.action = this.link.dataset.to!;
        form.style.display = "hidden";
        form.appendChild(
            this.#buildHiddenInput("_csrf", this.link.dataset.csrf!)
        );
        document.body.appendChild(form);
        submit.type = "submit";
        form.appendChild(submit);
        console.log(form);
        submit.click();
    }

    #buildHiddenInput(name: string, value: string) {
        var input = document.createElement("input");
        input.type = "hidden";
        input.name = name;
        input.value = value;
        return input;
    }
}

declare global {
    interface Window {
        SendFormElement: typeof SendFormElement
    }
}

if (!window.customElements.get('send-form')) {
    window.SendFormElement = SendFormElement
    window.customElements.define('send-form', SendFormElement)
}