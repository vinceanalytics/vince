!function(){"use strict";var e,t,n,s=window.location,w=window.document,d=w.getElementById("vince"),u=d.getAttribute("data-api")||(e=d.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function p(e){console.warn("Ignoring Event: "+e)}function i(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(s.hostname)||"file:"===s.protocol)return p("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return p("localStorage flag")}catch(e){}var n=d&&d.getAttribute("data-include"),i=d&&d.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(l),r=i&&i.split(",").some(l);if(!a||r)return p("exclusion rule")}var o={};o.n=e,o.u=t&&t.u?t.u:s.href,o.d=d.getAttribute("data-domain"),o.r=w.referrer||null,o.w=window.innerWidth,t&&t.meta&&(o.m=JSON.stringify(t.meta)),t&&t.props&&(o.p=t.props),o.h=1;var c=new XMLHttpRequest;c.open("POST",u,!0),c.setRequestHeader("Content-Type","text/plain"),c.send(JSON.stringify(o)),c.onreadystatechange=function(){4===c.readyState&&t&&t.callback&&t.callback()}}function l(e){var t=s.pathname;return(t+=s.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var a=window.vince&&window.vince.q||[];window.vince=i;for(var r=0;r<a.length;r++)i.apply(this,a[r])}();