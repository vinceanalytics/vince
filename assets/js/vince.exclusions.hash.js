!function(){"use strict";var s=window.location,d=window.document,w=d.currentScript,p=w.getAttribute("data-api")||new URL(w.src).origin+"/api/event";function u(e){console.warn("Ignoring Event: "+e)}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(s.hostname)||"file:"===s.protocol)return u("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return u("localStorage flag")}catch(e){}var n=w&&w.getAttribute("data-include"),i=w&&w.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(l),r=i&&i.split(",").some(l);if(!a||r)return u("exclusion rule")}var o={};o.n=e,o.u=s.href,o.d=w.getAttribute("data-domain"),o.r=d.referrer||null,o.w=window.innerWidth,t&&t.meta&&(o.m=JSON.stringify(t.meta)),t&&t.props&&(o.p=t.props),o.h=1;var c=new XMLHttpRequest;c.open("POST",p,!0),c.setRequestHeader("Content-Type","text/plain"),c.send(JSON.stringify(o)),c.onreadystatechange=function(){4===c.readyState&&t&&t.callback&&t.callback()}}function l(e){var t=s.pathname;return(t+=s.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,i=0;i<t.length;i++)e.apply(this,t[i]);function a(){n=s.pathname,e("pageview")}window.addEventListener("hashchange",a),"prerender"===d.visibilityState?d.addEventListener("visibilitychange",function(){n||"visible"!==d.visibilityState||a()}):a()}();