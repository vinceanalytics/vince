!function(){"use strict";var o=window.location,l=window.document,s=l.currentScript,c=s.getAttribute("data-api")||new URL(s.src).origin+"/api/event";function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var i=s&&s.getAttribute("data-include"),a=s&&s.getAttribute("data-exclude");if("pageview"===e){i=!i||i.split(",").some(n),a=a&&a.split(",").some(n);if(!i||a)return u("exclusion rule",t)}function n(e){var t=o.pathname;return(t+=o.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var i={},r=(i.n=e,i.u=o.href,i.d=s.getAttribute("data-domain"),i.r=l.referrer||null,t&&t.meta&&(i.m=JSON.stringify(t.meta)),t&&t.props&&(i.p=t.props),i.h=1,new XMLHttpRequest);r.open("POST",c,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(i)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var i,a=0;a<t.length;a++)e.apply(this,t[a]);function n(){i=o.pathname,e("pageview")}window.addEventListener("hashchange",n),"prerender"===l.visibilityState?l.addEventListener("visibilitychange",function(){i||"visible"!==l.visibilityState||n()}):n()}();