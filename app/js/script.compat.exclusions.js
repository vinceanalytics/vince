!function(){"use strict";var t,l=window.location,o=window.document,s=o.getElementById("plausible"),p=s.getAttribute("data-api")||(t=(t=s).src.split("/"),d=t[0],t=t[2],d+"//"+t+"/api/event");function u(t,e){t&&console.warn("Ignoring Event: "+t),e&&e.callback&&e.callback()}function e(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return u("localhost",e);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,e);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",e)}catch(t){}var a=s&&s.getAttribute("data-include"),i=s&&s.getAttribute("data-exclude");if("pageview"===t){a=!a||a.split(",").some(n),i=i&&i.split(",").some(n);if(!a||i)return u("exclusion rule",e)}function n(t){return l.pathname.match(new RegExp("^"+t.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},r=(a.n=t,a.u=l.href,a.d=s.getAttribute("data-domain"),a.r=o.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),new XMLHttpRequest);r.open("POST",p,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&e&&e.callback&&e.callback({status:r.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=e;for(var i,n=0;n<a.length;n++)e.apply(this,a[n]);function r(){i!==l.pathname&&(i=l.pathname,e("pageview"))}var c,d=window.history;d.pushState&&(c=d.pushState,d.pushState=function(){c.apply(this,arguments),r()},window.addEventListener("popstate",r)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){i||"visible"!==o.visibilityState||r()}):r()}();