!function(){"use strict";var l=window.location,r=window.document,o=r.currentScript,i=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var n={},a=(n.n=e,n.u=t&&t.u?t.u:l.href,n.d=o.getAttribute("data-domain"),n.r=r.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),n.h=1,new XMLHttpRequest);a.open("POST",i,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(n)),a.onreadystatechange=function(){4===a.readyState&&t&&t.callback&&t.callback({status:a.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n=0;n<t.length;n++)e.apply(this,t[n]);var s=1;function a(e){var t,n,a,r,o;function i(){a||(a=!0,window.location=n.href)}"auxclick"===e.type&&e.button!==s||((t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target))&&t.href&&t.href.split("?")[0],(o=t)&&o.href&&o.host&&o.host!==l.host&&(o=e,e={name:"Outbound Link: Click",props:{url:(n=t).href}},a=!1,!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(o,n)?((r={props:e.props}).revenue=e.revenue,plausible(e.name,r)):((r={props:e.props,callback:i}).revenue=e.revenue,plausible(e.name,r),setTimeout(i,5e3),o.preventDefault())))}r.addEventListener("click",a),r.addEventListener("auxclick",a)}();