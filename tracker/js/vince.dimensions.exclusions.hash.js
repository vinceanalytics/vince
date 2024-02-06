(function(){"use strict";var s,a,c,t=window.location,n=window.document,e=n.currentScript,l=e.getAttribute("data-api")||d(e);function o(e){console.warn("Ignoring Event: "+e)}function d(e){return new URL(e.src).origin+"/api/event"}function i(s,i){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(t.hostname)||t.protocol==="file:")return o("localhost");if(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)return;try{if(window.localStorage.vince_ignore==="true")return o("localStorage flag")}catch{}var a,r,c,h,m,p,d=e&&e.getAttribute("data-include"),u=e&&e.getAttribute("data-exclude");if(s==="pageview"&&(h=!d||d&&d.split(",").some(f),m=u&&u.split(",").some(f),!h||m))return o("exclusion rule");function f(e){var n=t.pathname;return n+=t.hash,n.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^.])\*/g,"$1[^\\s/]*")+"/?$"))}a={},a.n=s,a.u=t.href,a.d=e.getAttribute("data-domain"),a.r=n.referrer||null,a.w=window.innerWidth,i&&i.meta&&(a.m=JSON.stringify(i.meta)),i&&i.props&&(a.p=i.props),p=e.getAttributeNames().filter(function(e){return e.substring(0,6)==="event-"}),c=a.p||{},p.forEach(function(t){var n=t.replace("event-",""),s=e.getAttribute(t);c[n]=c[n]||s}),a.p=c,a.h=1,r=new XMLHttpRequest,r.open("POST",l,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){r.readyState===4&&i&&i.callback&&i.callback()}}a=window.vince&&window.vince.q||[],window.vince=i;for(s=0;s<a.length;s++)i.apply(this,a[s]);function r(){c=t.pathname,i("pageview")}window.addEventListener("hashchange",r);function u(){!c&&n.visibilityState==="visible"&&r()}n.visibilityState==="prerender"?n.addEventListener("visibilitychange",u):r()})()