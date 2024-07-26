!function(){"use strict";var l=window.location,o=window.document,s=o.currentScript,u=s.getAttribute("data-api")||new URL(s.src).origin+"/api/event";function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return p("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return p(null,t);try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var a=s&&s.getAttribute("data-include"),n=s&&s.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(i),n=n&&n.split(",").some(i);if(!a||n)return p("exclusion rule",t)}function i(e){return l.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},r=(a.n=e,a.u=l.href,a.d=s.getAttribute("data-domain"),a.r=o.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),new XMLHttpRequest);r.open("POST",u,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function i(){a!==l.pathname&&(a=l.pathname,e("pageview"))}var r,c=window.history;c.pushState&&(r=c.pushState,c.pushState=function(){r.apply(this,arguments),i()},window.addEventListener("popstate",i)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){a||"visible"!==o.visibilityState||i()}):i();var d=1;function f(e){var t,a,n,i,r;function o(){n||(n=!0,window.location=a.href)}"auxclick"===e.type&&e.button!==d||((t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target))&&t.href&&t.href.split("?")[0],(r=t)&&r.href&&r.host&&r.host!==l.host&&(r=e,e={name:"Outbound Link: Click",props:{url:(a=t).href}},n=!1,!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(r,a)?((i={props:e.props}).revenue=e.revenue,plausible(e.name,i)):((i={props:e.props,callback:o}).revenue=e.revenue,plausible(e.name,i),setTimeout(o,5e3),r.preventDefault())))}o.addEventListener("click",f),o.addEventListener("auxclick",f)}();