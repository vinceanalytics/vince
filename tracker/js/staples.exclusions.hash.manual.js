(function(){"use strict";var n,o,t=window.location,i=window.document,e=i.currentScript,r=e.getAttribute("data-api")||c(e);function s(e){console.warn("Ignoring Event: "+e)}function c(e){return new URL(e.src).origin+"/api/event"}function a(n,o){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(t.hostname)||t.protocol==="file:")return s("localhost");if(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)return;try{if(window.localStorage.vince_ignore==="true")return s("localStorage flag")}catch{}var a,c,u,h,l=e&&e.getAttribute("data-include"),d=e&&e.getAttribute("data-exclude");if(n==="pageview"&&(u=!l||l&&l.split(",").some(m),h=d&&d.split(",").some(m),!u||h))return s("exclusion rule");function m(e){var n=t.pathname;return n+=t.hash,n.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^.])\*/g,"$1[^\\s/]*")+"/?$"))}a={},a.n=n,a.u=o&&o.u?o.u:t.href,a.d=e.getAttribute("data-domain"),a.r=i.referrer||null,a.w=window.innerWidth,o&&o.meta&&(a.m=JSON.stringify(o.meta)),o&&o.props&&(a.p=o.props),a.h=1,c=new XMLHttpRequest,c.open("POST",r,!0),c.setRequestHeader("Content-Type","text/plain"),c.send(JSON.stringify(a)),c.onreadystatechange=function(){c.readyState===4&&o&&o.callback&&o.callback()}}o=window.vince&&window.vince.q||[],window.vince=a;for(n=0;n<o.length;n++)a.apply(this,o[n])})()