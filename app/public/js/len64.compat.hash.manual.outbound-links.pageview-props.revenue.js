!function(){"use strict";var e,t,l=window.location,i=window.document,o=i.getElementById("plausible"),u=o.getAttribute("data-api")||(e=(e=o).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function n(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return s("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return s(null,t);try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var n={},e=(n.n=e,n.u=t&&t.u?t.u:l.href,n.d=o.getAttribute("data-domain"),n.r=i.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),o.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),a=n.p||{},r=(e.forEach(function(e){var t=e.replace("event-",""),e=o.getAttribute(e);a[t]=a[t]||e}),n.p=a,n.h=1,new XMLHttpRequest);r.open("POST",u,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(n)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=n;for(var r=0;r<a.length;r++)n.apply(this,a[r]);var c=1;function p(e){var t,n,a,r,i;function o(){a||(a=!0,window.location=n.href)}"auxclick"===e.type&&e.button!==c||((t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target))&&t.href&&t.href.split("?")[0],(i=t)&&i.href&&i.host&&i.host!==l.host&&(i=e,e={name:"Outbound Link: Click",props:{url:(n=t).href}},a=!1,!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(i,n)?((r={props:e.props}).revenue=e.revenue,plausible(e.name,r)):((r={props:e.props,callback:o}).revenue=e.revenue,plausible(e.name,r),setTimeout(o,5e3),i.preventDefault())))}i.addEventListener("click",p),i.addEventListener("auxclick",p)}();