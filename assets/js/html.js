"use strict";

(function() {
  var PolyfillEvent = eventConstructor();

  function eventConstructor() {
    if (typeof window.CustomEvent === "function") return window.CustomEvent;
    // IE<=9 Support
    function CustomEvent(event, params) {
      params = params || {bubbles: false, cancelable: false, detail: undefined};
      var evt = document.createEvent('CustomEvent');
      evt.initCustomEvent(event, params.bubbles, params.cancelable, params.detail);
      return evt;
    }
    CustomEvent.prototype = window.Event.prototype;
    return CustomEvent;
  }

  function buildHiddenInput(name, value) {
    var input = document.createElement("input");
    input.type = "hidden";
    input.name = name;
    input.value = value;
    return input;
  }

  function handleClick(element, targetModifierKey) {
    var to = element.getAttribute("data-to"),
        method = buildHiddenInput("_method", element.getAttribute("data-method")),
        csrf = buildHiddenInput("_csrf_token", element.getAttribute("data-csrf")),
        form = document.createElement("form"),
        submit = document.createElement("input"),
        target = element.getAttribute("target");

    form.method = (element.getAttribute("data-method") === "get") ? "get" : "post";
    form.action = to;
    form.style.display = "hidden";

    if (target) form.target = target;
    else if (targetModifierKey) form.target = "_blank";

    form.appendChild(csrf);
    form.appendChild(method);
    document.body.appendChild(form);

    // Insert a button and click it instead of using `form.submit`
    // because the `submit` function does not emit a `submit` event.
    submit.type = "submit";
    form.appendChild(submit);
    submit.click();
  }

  window.addEventListener("click", function(e) {
    var element = e.target;
    if (e.defaultPrevented) return;

    while (element && element.getAttribute) {
      var vinceLinkEvent = new PolyfillEvent('vince.link.click', {
        "bubbles": true, "cancelable": true
      });

      if (!element.dispatchEvent(vinceLinkEvent)) {
        e.preventDefault();
        e.stopImmediatePropagation();
        return false;
      }

      if (element.getAttribute("data-method")) {
        handleClick(element, e.metaKey || e.shiftKey);
        e.preventDefault();
        return false;
      } else {
        element = element.parentNode;
      }
    }
  }, false);

  window.addEventListener('vince.link.click', function (e) {
    var message = e.target.getAttribute("data-confirm");
    if(message && !window.confirm(message)) {
      e.preventDefault();
    }
  }, false);
})();